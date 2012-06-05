package main

import (
	"builder"
	"encoding/gob"
	"log"
	"net/http"
)

func init() {
	gob.Register(&TestWork{})
}

type TestWork struct {
	FRevisions  []string
	vcs         builder.VCS
	FRepoPath   string
	FImportPath string
	Workspace   bool
}

func (t *TestWork) Revisions() []string { return t.FRevisions }
func (t *TestWork) VCS() builder.VCS    { return t.vcs }
func (t *TestWork) RepoPath() string    { return t.FRepoPath }
func (t *TestWork) ImportPath() string  { return t.FImportPath }
func (t *TestWork) IsWorkspace() bool   { return t.Workspace }

var test_work = &TestWork{
	FRevisions: []string{
		"e4ef402bacb2a4e0a86c0729ffd531e52eb68d52", //empty tests
		"1351a526989eda49cf7159561f38d9454c8e961a", //before go1 commit
		"c97d0b46f86c1d1294b9351c01349177e38ef2b3", //working with tests
		"b1a6b6797e2009e1dac7ccd5515f8aee17df6774", //tests that fail to compile
	},
	vcs:         builder.Git,
	FRepoPath:   "git://github.com/zeebo/irc",
	FImportPath: "github.com/zeebo/irc",
	Workspace:   false,
}

//our basic handle index that demonstrates how to get data from the context
//inside a template
func handle_simple_work(w http.ResponseWriter, req *http.Request, ctx *Context) {
	work_queue <- test_work
	log.Println("sent item in")
}
