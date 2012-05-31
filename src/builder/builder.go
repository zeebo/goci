package builder

import (
	"io/ioutil"
	"os"
	fp "path/filepath"
	"time"
)

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
	Path() string
	Clean() error
	Info() Info
}

type Info struct {
	Revision   string
	When       time.Time
	ImportPath string
	RepoPath   string
}

type build struct {
	path string
	err  error
	info *Info
}

func (b build) Info() Info {
	return *b.info
}

func (b build) Error() error {
	return b.err
}

func (b build) Path() string {
	return b.path
}

func (b build) Clean() (err error) {
	err = os.RemoveAll(fp.Base(b.path))
	return
}

var _ Build = build{}

func Build(w Work) (items []Build, err error) {
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
		bui := build{}

		//checkout the revision we need
		rep.Error = vcs.Checkout(tmpRepo, rev)
		if rep.Error != nil {
			rep.Duration = time.Since(rep.When)
			res = append(res, rep)
			continue
		}

		//copy the repo to the srcDir
		rep.Error = copy(tmpRepo+string(fp.Separator)+".", srcDir)
		if rep.Error != nil {
			rep.Duration = time.Since(rep.When)
			res = append(res, rep)
			continue
		}

		//figure out what packages need to be built
		packs, rep.Error = list(gopath)
		if rep.Error != nil {
			rep.Duration = time.Since(rep.When)
			res = append(res, rep)
			continue
		}

		//run a get to build deps
		rep.Error = get(gopath, packs...)
		if rep.Error != nil {
			rep.Duration = time.Since(rep.When)
			res = append(res, rep)
			continue
		}

		//build the binary and move it to a temporary directory
		rep.Output, rep.Passed, rep.Error = test(gopath, packs...)
		rep.Duration = time.Since(rep.When)
		res = append(res, rep)
	}

	return
}
