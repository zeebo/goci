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
	"os"
	"strings"
	"time"
	fp "path/filepath"
)

var (
	//Symlinks are unspported. Compress and Extract return this error if one is
	//encountered.
	ErrNoSymlinks     = errors.New("symlinks are unsupported")
	ErrInvalidExtract = errors.New("exctract attempted to go above directory")
)

//Compress takes the given directory and compresses it into a file written to
//out for later extraction.
func Compress(dir string, out io.Writer) (err error) {
	g := gzip.NewWriter(out)
	defer g.Close()

	t := tar.NewWriter(g)
	defer t.Close()

	dir = fp.Clean(dir)
	ds := "." + string(fp.Separator)

	err = walk(dir, func(path string, fi os.FileInfo, e error) (err error) {
		if e != nil {
			err = e
			return
		}

		//fix the path for the filename
		//TODO(zeebo): make sure this isn't a giant stupid hack
		fname, err := fp.Rel(dir, path)
		if err != nil {
			return
		}
		if fname != "." && !strings.HasPrefix(fname, ds) {
			fname = ds + fname
		}

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
func Extract(in io.Reader, dir string) (err error) {
	g, err := gzip.NewReader(in)
	if err != nil {
		return
	}
	defer g.Close()

	t := tar.NewReader(g)

	var of io.WriteCloser
	for {
		//grab the next file
		hdr, er := t.Next()
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			return
		}

		//make sure we don't go above our directory
		path := fp.Clean(fp.Join(dir, hdr.Name))
		if !strings.HasPrefix(path, dir) {
			err = ErrInvalidExtract
			return
		}

		//if it's a directory, try to make it
		if hdr.Typeflag == tar.TypeDir {
			err = world.MkdirAll(path)
			if err != nil {
				return
			}

			//continue on now
			continue
		}

		//create the file and copy the data into it
		of, err = world.Create(path)
		if err != nil {
			return
		}
		_, err = io.Copy(of, t)
		of.Close()

		if err != nil {
			return
		}
	}

	return
}

func headerFor(path string, fi os.FileInfo) (hdr *tar.Header, err error) {
	hdr = new(tar.Header)

	//only set the header fields we care about
	hdr.Name = path
	hdr.Mode = 0777
	hdr.ModTime = time.Now()

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
