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

type Report struct {
	When        time.Time
	Duration    time.Duration
	Revision    string
	Passed      bool
	Output      string
	Error       error  `json:"-"`
	ErrorString string `json:"Error"`
}

func Run(w Work) (res []Report, err error) {
	//create a gopath to run all this stuff in
	gopath, err := ioutil.TempDir("", "gopath")
	if err != nil {
		return
	}
	defer os.RemoveAll(gopath)

	if w.IsWorkspace() {
		res, err = cloneAndTest(w, gopath, gopath)
	} else {
		res, err = cloneAndTest(w, gopath, fp.Join(gopath, "src", w.ImportPath()))
	}

	//set the strings on the results
	for _, r := range res {
		if r.Error != nil {
			r.ErrorString = r.Error.Error()
		}
	}
	return
}

func cloneAndTest(w Work, gopath, srcDir string) (res []Report, err error) {
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
		rep := Report{
			When:     time.Now(),
			Revision: rev,
		}

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

		//run the tests
		rep.Output, rep.Passed, rep.Error = test(gopath, packs...)
		rep.Duration = time.Since(rep.When)
		res = append(res, rep)
	}

	return
}
