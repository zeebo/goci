package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

//our basic handle index that demonstrates how to get data from the context
//inside a template
func handle_index(w http.ResponseWriter, req *http.Request, ctx *Context) {
	if req.URL.Path != "/" {
		perform_status(w, req, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-type", "text/html")
	ctx.Set("content", "Content!")
	execute(w, ctx, tmpl_root("blocks", "index.block"))
}

//make a silly handler for testing statuses
func handle_status(w http.ResponseWriter, req *http.Request, ctx *Context) {
	status, err := strconv.ParseInt(path.Base(req.URL.Path), 10, 32)
	if err != nil {
		log.Println(err)
		perform_status(w, req, http.StatusNotFound)
		return
	}
	perform_status(w, req, int(status))
}

func handle_cmd(w http.ResponseWriter, req *http.Request, ctx *Context) {
	cmd := req.FormValue("cmd")
	if cmd == "" {
		perform_status(w, req, http.StatusNotFound)
		return
	}

	parts := strings.Split(cmd, " ")
	var ebuf, obuf bytes.Buffer
	ex := exec.Command(parts[0], parts[1:]...)
	ex.Stderr = &ebuf
	ex.Stdout = &obuf
	err := ex.Run()

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, "ebuf:", ebuf.String())
	fmt.Fprintln(w, "obuf:", obuf.String())
	fmt.Fprintln(w, "err:", err)
}
