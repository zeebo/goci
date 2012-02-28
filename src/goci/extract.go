package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func Extract(url string, path string) (err error) {
	logger.Println("Extracting", url, "to", path)
	defer logger.Println("Finished extracting.")

	res, err := http.Get(url)
	if err != nil {
		return
	}
	defer res.Body.Close()
	r, err := gzip.NewReader(res.Body)
	if err != nil {
		return
	}
	defer r.Close()

	err = untar(r, path)
	if err != nil {
		return
	}

	//check that the go bin is executable
	exe := filepath.Join(path, "bin", "go")
	stat, err := os.Stat(exe)
	if err != nil {
		return err
	}

	if m := stat.Mode(); m.IsDir() {
		return fmt.Errorf("%s is not executable.", exe)
	}

	return
}

func untar(r io.Reader, path string) error {
	tr := tar.NewReader(r)
	mode := os.O_RDWR | os.O_CREATE | os.O_TRUNC

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		//drop of the prefixing go/
		name := hdr.Name[3:]
		path := filepath.Join(path, name)

		//check if it's a directory
		if hdr.Typeflag == tar.TypeDir {
			if err := os.Mkdir(path, 0777); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, mode, 0777)
		if err != nil {
			return err
		}
		io.Copy(file, tr)
		file.Close()
	}
	return nil
}
