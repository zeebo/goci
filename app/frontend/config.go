package frontend

import (
	"errors"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

//Conf represents the configuration for the frontend.
type Conf struct {
	Templates string //the path to the templates
	Static    string //the path to static assets
	Debug     bool   //if true templates will be compiled on every invocation
}

//Config is the configuration for the frontend.
var Config = new(Conf)

//Open makes Config an http.FileServer for static files.
func (c *Conf) Open(name string) (http.File, error) {
	if filepath.Separator != '/' && strings.IndexRune(name, filepath.Separator) >= 0 {
		return nil, errors.New("http: invalid character in file path")
	}
	dir := c.Static
	if dir == "" {
		dir = "."
	}
	f, err := os.Open(filepath.Join(dir, filepath.FromSlash(path.Clean("/"+name))))
	if err != nil {
		return nil, err
	}
	return f, nil
}
