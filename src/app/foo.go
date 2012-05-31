package main

import (
	"builder"
	"log"
	"net/http"
)

type testWork struct {
	revisions  []string
	vcs        builder.VCS
	repoPath   string
	importPath string
	workspace  bool
}

func (t *testWork) Revisions() []string { return t.revisions }
func (t *testWork) VCS() builder.VCS    { return t.vcs }
func (t *testWork) RepoPath() string    { return t.repoPath }
func (t *testWork) ImportPath() string  { return t.importPath }
func (t *testWork) IsWorkspace() bool   { return t.workspace }

var test_work = &testWork{
	revisions: []string{
		"e4ef402bacb2a4e0a86c0729ffd531e52eb68d52", //empty tests
		"1351a526989eda49cf7159561f38d9454c8e961a", //before go1 commit
		"c97d0b46f86c1d1294b9351c01349177e38ef2b3", //working with tests
		"b1a6b6797e2009e1dac7ccd5515f8aee17df6774", //tests that fail to compile
	},
	vcs:        builder.Git,
	repoPath:   "git://github.com/zeebo/irc",
	importPath: "github.com/zeebo/irc",
	workspace:  false,
}

//our basic handle index that demonstrates how to get data from the context
//inside a template
func handle_simple_work(w http.ResponseWriter, req *http.Request, ctx *Context) {
	work_queue <- test_work
	log.Println("sent item in")
}
