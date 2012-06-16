package main

import (
	"builder"
	"encoding/gob"
	"log"
	"net/http"
	"worker"
)

func init() {
	gob.Register(&TestWork{})
}

type TestWork struct {
	FRevisions  []string
	FVCS        builder.VCS
	FRepoPath   string
	FImportPath string
	Workspace   bool
	FLink       string
}

func (t *TestWork) Revisions() []string { return t.FRevisions }
func (t *TestWork) VCS() builder.VCS    { return t.FVCS }
func (t *TestWork) RepoPath() string    { return t.FRepoPath }
func (t *TestWork) ImportPath() string  { return t.FImportPath }
func (t *TestWork) IsWorkspace() bool   { return t.Workspace }
func (t *TestWork) Link() string        { return t.FLink }

var test_work = &TestWork{
	FRevisions: []string{
		"e4ef402bacb2a4e0a86c0729ffd531e52eb68d52", //empty tests
		"1351a526989eda49cf7159561f38d9454c8e961a", //before go1 commit
		"c97d0b46f86c1d1294b9351c01349177e38ef2b3", //working with tests
		"b1a6b6797e2009e1dac7ccd5515f8aee17df6774", //tests that fail to compile
	},
	FVCS:        builder.Git,
	FRepoPath:   "git://github.com/zeebo/irc",
	FImportPath: "github.com/zeebo/irc",
	Workspace:   false,
	FLink:       "http://github.com/zeebo/irc",
}

//our basic handle index that demonstrates how to get data from the context
//inside a template
func handle_simple_work(w http.ResponseWriter, req *http.Request, ctx *Context) {
	worker.Schedule(test_work)
	log.Println("sent item in")
}
