package builder

import (
	"io/ioutil"
	"os"
	"path"
	fp "path/filepath"
	"runtime"
)

var exeSuffix = func() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}()

//Work represents an item of work to be completed by the builder
type Work interface {
	Revisions() (rev []string)
	VCS() (v VCS)
	RepoPath() (path string)
	ImportPath() (path string)
	IsWorkspace() (ok bool)
}

type Build interface {
	Error() error
	Paths() []string
	Revision() string
	Cleanup() error
}

type build struct {
	paths []string
	base  string
	err   error
	rev   string
}

func (b build) Revision() string {
	return b.rev
}

func (b build) Error() error {
	return b.err
}

func (b build) Paths() []string {
	return b.paths
}

func (b build) Cleanup() (err error) {
	if b.base != "" {
		err = os.RemoveAll(b.base)
	}
	return
}

var _ Build = build{}

func CreateBuilds(w Work) (items []Build, err error) {
	//create a gopath to run all this stuff in
	gopath, err := ioutil.TempDir("", "gopath")
	if err != nil {
		return
	}
	defer os.RemoveAll(gopath)

	if w.IsWorkspace() {
		items, err = cloneAndTest(w, gopath, gopath)
	} else {
		items, err = cloneAndTest(w, gopath, fp.Join(gopath, "src", w.ImportPath()))
	}

	return
}

func cloneAndTest(w Work, gopath, srcDir string) (res []Build, err error) {
	vcs := w.VCS()

	tmpRepo, err := ioutil.TempDir("", "tmpRepo")
	if err != nil {
		return
	}
	defer os.RemoveAll(tmpRepo)

	//clone to a temporary location
	err = vcs.Clone(w.RepoPath(), tmpRepo)
	if err != nil {
		return
	}

	var packs []string
	for _, rev := range w.Revisions() {
		//clean bin/pkg directories from the gopath
		os.RemoveAll(fp.Join(gopath, "pkg"))
		os.RemoveAll(fp.Join(gopath, "bin"))

		bui := build{
			rev: rev,
		}

		//make a new directory for the builds of this revision
		bui.base, bui.err = ioutil.TempDir("", rev)
		if bui.err != nil {
			goto done
		}

		//checkout the revision we need
		bui.err = vcs.Checkout(tmpRepo, rev)
		if bui.err != nil {
			goto done
		}

		//copy the repo to the srcDir
		bui.err = copy(tmpRepo+string(fp.Separator)+".", srcDir)
		if bui.err != nil {
			goto done
		}

		//figure out what packages need to be built
		packs, bui.err = list(gopath)
		if bui.err != nil {
			goto done
		}

		//run a get to build deps
		bui.err = get(gopath, packs...)
		if bui.err != nil {
			goto done
		}

		//build the binaries and move them to a temporary directory
		for _, pack := range packs {
			bui.err = testbuild(gopath, pack, bui.base)
			if bui.err != nil {
				goto done
			}

			//what the go tool does from inspecting the source
			_, elem := path.Split(pack)
			name := elem + ".test" + exeSuffix
			path := fp.Join(bui.base, name)

			//make sure that binary exists before we add it to the paths. it may
			//not exist if there are no test files, so just continue past this
			//commit.
			if _, err := os.Stat(path); err != nil {
				continue
			}

			//append the path to the new dir into the paths
			bui.paths = append(bui.paths, path)
		}

	done:
		//if we didn't create any binaries, don't keep the dump directory around
		if len(bui.paths) == 0 {
			os.RemoveAll(bui.base)
			bui.base = ""
		}

		res = append(res, bui)
	}

	return
}
