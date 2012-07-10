package builder

import (
	"errors"
	"fmt"
	"github.com/zeebo/goci/environ"
	"os"
	"runtime"
	"time"
	fp "path/filepath"
	p "path"
)

type localWorld interface {
	Exists(string) bool
	LookPath(string) (string, error)
	TempDir(string) (string, error)
	Make(environ.Command) environ.Proc
}

var (
	ErrTooMany         = errors.New("too many revisions in that work item")
	ErrUnknownWorkType = errors.New("unknown work type")

	world localWorld = environ.New()
)

//Builder is a type that builds go packages at specified revisions.
type Builder struct {
	goos, goarch string
	goroot       string
	gopath       string

	//generated
	base_env []string
	env      []string
}

//New returns a Builder that can be used for building Work objects.
//If GOOS or GOARCH are the empty string, then the values from runtime are used.
//If GOROOT is the empty string, then value from the environment is used.
//If it is not set in the environment, New will panic.
//If GOPATH is the empty string, a temporary directory is created and used.
//If any error occurs creating the temporary directory, New will panic.
//If a directory has been created for the GOPATH, the Cleanup method will remove
//it.
//Commands run by the Builder use the PATH variable from the environment.
func New(GOOS, GOARCH, GOROOT string) (b Builder) {
	//fill in default values
	if GOOS == "" {
		GOOS = runtime.GOOS
	}
	if GOARCH == "" {
		GOARCH = runtime.GOARCH
	}
	if GOROOT == "" {
		GOROOT = must_env("GOROOT")
	}

	//create the builder
	b = Builder{
		goos:   GOOS,
		goarch: GOARCH,
		goroot: GOROOT,
		base_env: []string{
			fmt.Sprintf("GOROOT=%s", GOROOT),
			fmt.Sprintf("GOOS=%s", GOOS),
			fmt.Sprintf("GOARCH=%s", GOARCH),
			fmt.Sprintf("PATH=%s", must_env("PATH")),
		},
	}
	return
}

//Cleanup removes any temporary files created by the Builder. It is intended to
//be called after all work items the Builder will ever create have been created,
//like during the exit of a program.
func (b Builder) Cleanup() {
	os.RemoveAll(b.gopath)
}

//exeSuffix is a value that is appended to the end of a binary depending on what
//os we're using.
//TODO(zeebo): make sure this is right for cross platform builders
var exeSuffix = func() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}()

//Work is an incoming work item to generate the builds for a given revision and
//import path. If Revision is empty, the revision chosen by go get is used. If
//Subpackages is true, it will build binaries for all subpackages of the import
//path as well.
type Work struct {
	Revision    string
	ImportPath  string
	Subpackages bool

	//VCSHint is an optional parameter that specifies the version control system
	//used by the package. If set to the empty string, we will search for the 
	//system by looking for the metadata directory.
	VCSHint string
}

//Build is a type that represents a built test and tarballed source ready to be
//sent to a runner. It contains an error if there were any problems building the
//test binary.
type Build struct {
	Date time.Time

	BinaryPath string
	SourcePath string

	Error string
}

//Clean removes the directories that the binary and tarball are in.
func (b Build) Clean() {
	os.RemoveAll(p.Base(b.BinaryPath))
	os.RemoveAll(p.Base(b.SourcePath))
}

func (b Builder) Build(w *Work) (builds []Build, revDate time.Time, err error) {
	//create a GOPATH for this work item
	b.gopath, err = world.TempDir("gopath")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(b.gopath)

	//set up the env to include the new gopath
	b.env = append(b.env, b.base_env...)
	b.env = append(b.env, fmt.Sprintf("GOPATH=%s", b.gopath))

	//get the import path (just download the package)
	if err = b.goGet(true, w.ImportPath); err != nil {
		return
	}

	//we can find the downloaded package in the first entry of the gopath
	packDir := fp.Join(b.gopath, "src", w.ImportPath)

	//set up the vcs
	var v vcs

	//check the hint for the vcs and fallback on searching the directories
	if vc, ok := vcsMap[w.VCSHint]; ok {
		v = vc
	} else {
		v = findVcs(packDir)
	}

	//if we don't have a vcs then we can't continue
	if v == nil {
		err = fmt.Errorf("unable to determine vcs for %s", w.ImportPath)
		return
	}

	//if we have a revision specified then do a checkout, otherwise, find it
	if w.Revision != "" {
		if err = v.Checkout(packDir, w.Revision); err != nil {
			return
		}
	} else {
		w.Revision, err = v.Current(packDir)
		if err != nil {
			return
		}
	}

	//set the date for the revision
	if revDate, err = v.Date(packDir, w.Revision); err != nil {
		return
	}

	//list the import path to determine how many builds there will be and what
	//packages need to be installed for the tests to compile
	path := w.ImportPath
	if w.Subpackages {
		path = p.Join(path, "...")
	}
	paths, testpaths, err := b.goList(path)
	if err != nil {
		return
	}

	//make a uniqued copy of all the paths we're going to download and install
	deppaths := make([]string, 0, len(paths)+len(testpaths))
	deppaths = append(deppaths, paths...)
	deppaths = append(deppaths, testpaths...)
	deppaths = unique(deppaths)

	//download, update and install all the deps this revision imports
	if err = b.goGet(false, deppaths...); err != nil {
		return
	}

	//build each of the tests
	for _, tpath := range paths {
		bu := b.build(tpath)

		//cover all the cases to append the build.
		switch {
		case bu.Error == "" && world.Exists(bu.BinaryPath):
			builds = append(builds, bu)
		case bu.Error != "":
			builds = append(builds, bu)
		default:
			bu.Clean()
		}
	}

	return
}

//build generates a test binary and tarball of the source for a given import path
//and returns a Build that represents this data.
func (b Builder) build(path string) (bu Build) {
	var err error

	bu.Date = time.Now()
	bu.BinaryPath, err = b.goTest(path)
	if err != nil {
		bu.Error = err.Error()
		return
	}

	tardir, err := world.TempDir("tarball")
	if err != nil {
		bu.Error = err.Error()
		return
	}

	//we can find the downloaded package in the first entry of the gopath
	packDir := fp.Join(b.gopath, "src", path)

	bu.SourcePath = fp.Join(tardir, "src.tar.gz")
	if err = tarball(packDir, bu.SourcePath); err != nil {
		bu.Error = err.Error()
		return
	}

	return
}
