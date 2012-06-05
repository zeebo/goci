package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

//pervasive type. convenient to have a short name for it.
type d map[string]interface{}

func base_execute(w io.Writer, ctx interface{}, blocks ...string) (err error) {
	if err = base_template.Execute(w, ctx, blocks...); err != nil {
		log.Println(err)
	}
	return
}

func perform_status(w http.ResponseWriter, ctx *Context, status int) {
	w.WriteHeader(status)
	block := fmt.Sprintf(tmpl_root("status", "%d.block"), status)
	if err := base_template.Execute(w, ctx, block); err != nil {
		log.Println(err)
	}
}

func internal_error(w http.ResponseWriter, req *http.Request, ctx *Context, err error) {
	perform_status(w, ctx, http.StatusInternalServerError)
	log.Println("error serving request:", err)
}

func tmpl_root(path ...string) string {
	elems := append([]string{template_dir}, path...)
	return filepath.Join(elems...)
}

func asset_root(path ...string) string {
	elems := append([]string{assets_dir}, path...)
	return filepath.Join(elems...)
}

func env(key, def string) (k string) {
	if k = os.Getenv(key); k == "" {
		k = def
	}
	return
}

func need_env(key string) (val string) {
	if val = os.Getenv(key); val == "" {
		panic("key not found: " + key)
	}
	return
}

// Serves static files from filesystemDir when any request is made matching
// requestPrefix
func serve_static(requestPrefix, filesystemDir string) {
	fileServer := http.FileServer(http.Dir(filesystemDir))
	handler := http.StripPrefix(requestPrefix, fileServer)
	http.Handle(requestPrefix+"/", handler)
}

func new_id() string {
	const idSize = 10
	var (
		buf [idSize]byte
		n   int
	)
	for n < idSize {
		m, err := rand.Read(buf[n:])
		if err != nil {
			log.Panicf("error generating a random id [%d bytes of %d]: %v", n, idSize, err)
		}
		n += m
	}
	return fmt.Sprintf("%X", buf)
}
