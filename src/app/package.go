package main

import (
	"builder"
	"encoding/gob"
)

type Package struct {
	Import string
}

func (p Package) Revisions() (r []string)    { return }
func (p Package) VCS() (v builder.VCS)       { return }
func (p Package) RepoPath() (r string)       { return }
func (p Package) ImportPath() string         { return p.Import }
func (p Package) WorkType() builder.WorkType { return builder.WorkTypeGoinstall }
func (p Package) Link() (l string)           { return }

func init() {
	gob.Register(Package{})
}
