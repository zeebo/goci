package tarball

import (
	"errors"
	"github.com/zeebo/goci/environ"
	"io"
	"os"
	fp "path/filepath"
	"sort"
)

type localWorld interface {
	//get info about files/directories
	Stat(string) (os.FileInfo, error)
	Readdir(string) ([]os.FileInfo, error)

	//create/open files
	Create(string, os.FileMode) (io.WriteCloser, error)
	Open(string) (io.ReadCloser, error)
	MkdirAll(string, os.FileMode) error
}

var (
	world localWorld = environ.New()

	skipDir = errors.New("skip this directory")
)

type walkFunc func(string, os.FileInfo, error) error

func walk(root string, walkFn walkFunc) error {
	info, err := world.Stat(root)
	if err != nil {
		return walkFn(root, nil, err)
	}
	return walkRec(root, info, walkFn)
}

func walkRec(path string, info os.FileInfo, walkFn walkFunc) error {
	err := walkFn(path, info, nil)
	if err != nil {
		if info.IsDir() && err == skipDir {
			return nil
		}
		return err
	}

	if !info.IsDir() {
		return nil
	}

	list, err := world.Readdir(path)
	if err != nil {
		return walkFn(path, info, err)
	}
	sort.Sort(names(list))

	for _, fi := range list {
		err = walkRec(fp.Join(path, fi.Name()), fi, walkFn)
		if err != nil {
			if !fi.IsDir() || err != skipDir {
				return err
			}
		}
	}

	return nil
}

// byName implements sort.Interface.
type names []os.FileInfo

func (n names) Len() int           { return len(n) }
func (n names) Less(i, j int) bool { return n[i].Name() < n[j].Name() }
func (n names) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
