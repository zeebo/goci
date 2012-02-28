package main

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func Extract(url string, path string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	r, err := gzip.NewReader(res.Body)
	if err != nil {
		return err
	}
	defer r.Close()
	return untar(r, path)
}

func untar(r io.Reader, path string) error {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		//drop of the prefixing go/
		path := filepath.Join(path, hdr.Name[3:])

		//check if it's a directory
		if hdr.Mode&(1<<3) > 0 {
			if err := os.Mkdir(path, 0777); err != nil {
				return err
			}
			continue
		}

		file, err := os.Create(path)
		if err != nil {
			return err
		}
		io.Copy(file, tr)
		file.Close()
	}
	return nil
}
