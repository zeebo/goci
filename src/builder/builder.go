package builder

import (
	"encoding/gob"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"sort"
	"time"
	fp "path/filepath"
)

var (
	ErrTooMany         = errors.New("too many revisions in that work item")
	ErrUnknownWorkType = errors.New("unknown work type")
)

func init() {
	gob.Register(build{})
}

var exeSuffix = func() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}()

type WorkType int

const (
	WorkTypePackage WorkType = iota
	WorkTypeWorkspace
	WorkTypeGoinstall
)

//Work represents an item of work to be completed by the builder
type Work interface {
	Revisions() (rev []string)
	VCS() (v VCS)
	RepoPath() (path string)
	ImportPath() (path string)
	WorkType() (t WorkType)
}

type Build interface {
	Error() error
	Paths() []Bundle
	Revision() string
	Date() time.Time
	Cleanup() error
}

type build struct {
	Ps   []string
	Ts   []string
	base string
	Err  error

	Rev  string
	RevT time.Time
}

func (b build) Revision() string {
	return b.Rev
}

func (b build) Error() error {
	return b.Err
}

func (b build) Paths() (bu []Bundle) {
	var tarb string
	for i, v := range b.Ps {
		if i < len(b.Ts) {
			tarb = b.Ts[i]
		} else {
			tarb = "" //no tarball for this path
		}

		bu = append(bu, Bundle{
			Path:    v,
			Tarball: tarb,
		})
	}
	return
}

func (b build) Date() time.Time {
	return b.RevT
}

func (b build) Cleanup() (err error) {
	if b.base != "" {
		err = os.RemoveAll(b.base)
	}
	return
}

//make it sortable based on the revision time
type sortBuild []Build

func (b sortBuild) Len() int           { return len(b) }
func (b sortBuild) Less(i, j int) bool { return b[i].Date().Before(b[j].Date()) }
func (b sortBuild) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

func (bui *build) appendPath(pack, tarball string) {
	//what the go tool does from inspecting the source
	_, elem := path.Split(pack)
	name := elem + ".test" + exeSuffix
	path := fp.Join(bui.base, name)

	//make sure that binary exists before we add it to the paths. it may
	//not exist if there are no test files, so only add it if something
	//is there.
	if _, err := os.Stat(path); err == nil {
		bui.Ps = append(bui.Ps, path)
		bui.Ts = append(bui.Ts, tarball)
	}
}

var _ Build = build{}
var _ sort.Interface = sortBuild{}

type environ struct {
	gopath  string
	srcDir  string
	tmpRepo string
	vcs     VCS
}

func (e environ) Cleanup() {
	if e.gopath != "" {
		os.RemoveAll(e.gopath)
	}
	if e.tmpRepo != "" {
		os.RemoveAll(e.tmpRepo)
	}
}

func (e environ) CleanGopath() {
	//clean bin/pkg directories from the gopath
	if e.gopath != "" {
		os.RemoveAll(fp.Join(e.gopath, "pkg"))
		os.RemoveAll(fp.Join(e.gopath, "bin"))
		os.RemoveAll(fp.Join(e.gopath, "src"))
	}
}

func newEnviron(w Work) (e environ, err error) {
	e.vcs = w.VCS()

	e.gopath, err = ioutil.TempDir("", "gopath")
	if err != nil {
		return
	}

	e.tmpRepo, err = ioutil.TempDir("", "tmpRepo")
	if err != nil {
		return
	}

	switch w.WorkType() {
	case WorkTypePackage:
		e.srcDir = fp.Join(e.gopath, "src", w.ImportPath())
	case WorkTypeWorkspace:
		e.srcDir = e.gopath
	default:
		err = ErrUnknownWorkType
	}

	return
}

func CreateBuilds(w Work) (items []Build, err error) {
	//check if we have a goget thing here
	if w.WorkType() == WorkTypeGoinstall {
		var item build
		item, err = createGoinstallBuild(w)
		if err == nil {
			items = append(items, item)
		}
		return
	}

	if len(w.Revisions()) > 5 {
		err = ErrTooMany
		return
	}

	//create a new environment for the work
	e, err := newEnviron(w)
	defer e.Cleanup()
	if err != nil {
		return
	}

	//grab the build items
	items, err = createNormalBuilds(w, e)

	//sort them based on date (earliest first)
	sort.Sort(sortBuild(items))
	return
}

func createNormalBuilds(w Work, e environ) (res []Build, err error) {
	//clone the repo to a temporary location for checkout/copying
	err = e.vcs.Clone(w.RepoPath(), e.tmpRepo)
	if err != nil {
		return
	}

	for _, rev := range w.Revisions() {
		//create the build binaries for this revision
		bui := createNormalBuild(rev, e)

		//if we didn't create any binaries, don't keep the dump directory around
		if len(bui.Ps) == 0 {
			bui.Cleanup()
			bui.base = ""
		}

		//add the build into our result list
		res = append(res, bui)

		//clean up the gopath from the last build
		e.CleanGopath()
	}

	return
}

func createNormalBuild(rev string, e environ) (bui build) {
	var packs, testpacks []string
	bui.Rev = rev

	//make a new directory for the builds of this revision
	bui.base, bui.Err = ioutil.TempDir("", rev)
	if bui.Err != nil {
		return
	}

	//checkout the revision we need
	bui.Err = e.vcs.Checkout(e.tmpRepo, rev)
	if bui.Err != nil {
		return
	}

	//get the date on that revision
	bui.RevT, bui.Err = e.vcs.Date(e.tmpRepo, rev)
	if bui.Err != nil {
		return
	}

	//copy the repo to the srcDir
	bui.Err = copy(e.tmpRepo+string(fp.Separator)+".", e.srcDir)
	if bui.Err != nil {
		return
	}

	//figure out what packages need to be built
	packs, testpacks, bui.Err = list(e.gopath)
	if bui.Err != nil {
		return
	}

	merged := make([]string, 0, len(packs)+len(testpacks))
	merged = append(merged, packs...)
	merged = append(merged, testpacks...)

	//run a get to build deps
	bui.Err = get(e.gopath, false, merged...)
	if bui.Err != nil {
		return
	}

	//build the binaries
	for _, pack := range packs {
		//create the build binary
		bui.Err = testbuild(e.gopath, pack, bui.base)
		if bui.Err != nil {
			return
		}

		//create a tarball
		pack_src := fp.Join(e.gopath, "src", pack)
		tarb := fp.Join(bui.base, hash(pack))
		bui.Err = tarball(pack_src, tarb)
		if bui.Err != nil {
			return
		}

		bui.appendPath(pack, tarb)
	}

	return
}

func createGoinstallBuild(w Work) (bui build, err error) {
	bui.Rev = "Latest"

	pack := w.ImportPath()

	//make a new directory for the build
	bui.base, bui.Err = ioutil.TempDir("", hash(pack))
	if bui.Err != nil {
		return
	}

	bui.Err = get(GOPATH, true, pack)
	if bui.Err != nil {
		return
	}

	//check if we can get a revision
	src_dir := fp.Join(GOPATH, "src", pack)
	for _, v := range []VCS{Git, HG} {
		if r, err := v.Current(src_dir); err == nil {
			bui.Rev = r
			break
		}
	}

	//check if we can get a date, only if we have a revision
	if bui.Rev != "Latest" {
		for _, v := range []VCS{Git, HG} {
			if t, err := v.Date(src_dir, bui.Rev); err == nil {
				bui.RevT = t
				break
			}
		}
	}

	//find all the deps for the tests and build those
	var testpacks []string
	_, testpacks, bui.Err = listPackage(GOPATH, pack)
	if bui.Err != nil {
		return
	}

	if len(testpacks) > 0 {
		//get the code and update it
		bui.Err = get(GOPATH, true, testpacks...)
		if bui.Err != nil {
			return
		}
	}

	//create the build binary
	bui.Err = testbuild(GOPATH, pack, bui.base)
	if bui.Err != nil {
		return
	}

	//create a tarball
	tarb := fp.Join(bui.base, "src.tar.gz")
	bui.Err = tarball(src_dir, tarb)
	if bui.Err != nil {
		return
	}

	bui.appendPath(pack, tarb)
	return
}
