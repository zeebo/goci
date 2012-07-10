//package tarball allows you to tarball and untarball directories
//
//tarball allows you to take directories and compress them into .tar.gz files.
//Compress is equivelant to running the command:
//
//	tar -cvzf <out> -C <dir> .
//
//where Extract is equivelant to running the command:
//
//	tar zxf <in> -C <dir>
//
package tarball

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"log"
	"os"
	fp "path/filepath"
)

var (
	//Symlinks are unspported. Compress and Extract return this error if one is
	//encountered.
	ErrNoSymlinks = errors.New("symlinks are unsupported")
)

//Compress takes the given directory and compresses it into a file written to
//out for later extraction.
func Compress(dir, out string) (err error) {
	f, err := world.Create(out)
	if err != nil {
		return
	}
	defer f.Close()

	g := gzip.NewWriter(f)
	defer g.Close()

	t := tar.NewWriter(g)
	defer t.Close()

	dir = fp.Clean(dir)

	err = walk(dir, func(path string, fi os.FileInfo, e error) (err error) {
		if e != nil {
			err = e
			return
		}

		//fix the path for the filename
		//TODO(zeebo): make sure this isn't a giant stupid hack
		var fname string
		if dir == "." {
			if path == "." {
				fname = "."
			} else {
				fname = "./" + path
			}
		} else {
			fname = "." + path[len(dir):]
		}
		log.Printf("a %s", fname)

		//create the header for the file (strip the directory out)
		hdr, err := headerFor(fname, fi)
		if err != nil {
			return
		}

		//write the header for this file
		if err = t.WriteHeader(hdr); err != nil {
			return
		}

		//don't need to copy the file in if it's a directory
		if fi.IsDir() {
			return
		}

		//open up the file for reading
		r, err := world.Open(path)
		if err != nil {
			return
		}
		defer r.Close()

		//copy the data in
		_, err = io.Copy(t, io.LimitReader(r, hdr.Size))

		return
	})

	return
}

//Extract takes a compressed file and extracts it to the given directory.
func Extract(in, dir string) (err error) {
	return
}

func headerFor(path string, fi os.FileInfo) (hdr *tar.Header, err error) {
	hdr = new(tar.Header)

	//only set the header fields we care about
	hdr.Name = path
	hdr.Mode = 0777

	//figure out what kind of file we have
	switch mode := fi.Mode(); {
	case mode&os.ModeDir == os.ModeDir:
		hdr.Typeflag = tar.TypeDir
	case mode&os.ModeSymlink == os.ModeSymlink:
		err = ErrNoSymlinks
		return
	default: //assume a regular file! go hog wild!
		hdr.Typeflag = tar.TypeReg
		hdr.Size = fi.Size()
	}

	return
}
