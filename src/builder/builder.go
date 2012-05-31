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

type environ struct {
	gopath  string
	srcDir  string
	tmpRepo string
	vcs     VCS
}

func CreateBuilds(w Work) (items []Build, err error) {
	//create a gopath to run all this stuff in
	gopath, err := ioutil.TempDir("", "gopath")
	if err != nil {
		return
	}
	defer os.RemoveAll(gopath)

	//set up an environment for the builds
	e := environ{
		gopath: gopath,
		vcs:    w.VCS(),
	}

	//set the srcdir of the environment
	if w.IsWorkspace() {
		e.srcDir = gopath
	} else {
		e.srcDir = fp.Join(gopath, "src", w.ImportPath())
	}

	//grab the build items
	items, err = createBuilds(w, e)
	return
}

func createBuilds(w Work, e environ) (res []Build, err error) {
	e.tmpRepo, err = ioutil.TempDir("", "tmpRepo")
	if err != nil {
		return
	}
	defer os.RemoveAll(e.tmpRepo)

	//clone to a temporary location
	err = e.vcs.Clone(w.RepoPath(), e.tmpRepo)
	if err != nil {
		return
	}

	for _, rev := range w.Revisions() {
		//clean bin/pkg directories from the gopath
		os.RemoveAll(fp.Join(e.gopath, "pkg"))
		os.RemoveAll(fp.Join(e.gopath, "bin"))

		//create the build binaries for this revision
		bui := createBuild(rev, e)

		//if we didn't create any binaries, don't keep the dump directory around
		if len(bui.paths) == 0 {
			os.RemoveAll(bui.base)
			bui.base = ""
		}

		res = append(res, bui)
	}

	return
}

func createBuild(rev string, e environ) (bui build) {
	var packs []string
	bui.rev = rev

	//make a new directory for the builds of this revision
	bui.base, bui.err = ioutil.TempDir("", rev)
	if bui.err != nil {
		return
	}

	//checkout the revision we need
	bui.err = e.vcs.Checkout(e.tmpRepo, rev)
	if bui.err != nil {
		return
	}

	//copy the repo to the srcDir
	bui.err = copy(e.tmpRepo+string(fp.Separator)+".", e.srcDir)
	if bui.err != nil {
		return
	}

	//figure out what packages need to be built
	packs, bui.err = list(e.gopath)
	if bui.err != nil {
		return
	}

	//run a get to build deps
	bui.err = get(e.gopath, packs...)
	if bui.err != nil {
		return
	}

	//build the binaries and move them to a temporary directory
	for _, pack := range packs {
		bui.err = testbuild(e.gopath, pack, bui.base)
		if bui.err != nil {
			return
		}

		//what the go tool does from inspecting the source
		_, elem := path.Split(pack)
		name := elem + ".test" + exeSuffix
		path := fp.Join(bui.base, name)

		//make sure that binary exists before we add it to the paths. it may
		//not exist if there are no test files, so only add it if something
		//is there.
		if _, err := os.Stat(path); err == nil {
			bui.paths = append(bui.paths, path)
		}
	}

	return
}
