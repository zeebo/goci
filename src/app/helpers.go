package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

//pervasive type. convenient to have a short name for it.
type d map[string]interface{}

//lists of things with easy append/duplicate methdos
type list []string

func (l *list) Append(val string) {
	*l = append(*l, val)
}

func (l list) Dup() (r list) {
	r = make(list, 0, len(l))
	for _, v := range l {
		r = append(r, v)
	}
	return
}

//perform_status runs the passed in status on the request and calls the appropriate block
func perform_status(w http.ResponseWriter, req *http.Request, status int) (err error) {
	w.WriteHeader(status)
	block := fmt.Sprintf(tmpl_root("status", "%d.block"), status)
	err = base_template.Execute(w, init_context(req), block)
	if err != nil {
		log.Println(err)
	}
	return
}

//execute is a convenient shorthand all it does extra is log errors and make the code
//a little cleaner
func execute(w http.ResponseWriter, ctx interface{}, blocks ...string) (err error) {
	if err = base_template.Execute(w, ctx, blocks...); err != nil {
		log.Println(err)
	}
	return
}

//internal_error is what is called when theres an error processing something
func internal_error(w http.ResponseWriter, req *http.Request, err error) {
	perform_status(w, req, http.StatusInternalServerError)
	log.Println("error serving request:", err)
}

func serve_static(requestPrefix, filesystemDir string) {
	fileServer := http.FileServer(http.Dir(filesystemDir))
	handler := http.StripPrefix(requestPrefix, fileServer)
	http.Handle(requestPrefix+"/", handler)
}

func tmpl_root(path ...string) string {
	elems := append([]string{template_dir}, path...)
	return filepath.Join(elems...)
}

func asset_root(path ...string) string {
	elems := append([]string{assets_dir}, path...)
	return filepath.Join(elems...)
}

func env(key, def string) string {
	if k := os.Getenv(key); k != "" {
		return k
	}
	return def
}
