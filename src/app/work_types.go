package main

import (
	"builder"
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"time"
)

type TaskInfo struct {
	When  time.Time
	ID    string `bson:"_id"`
	Error string
}

func (t TaskInfo) GetInfo() TaskInfo {
	return t
}

type Work struct {
	TaskInfo `bson:",inline"`
	Work     builder.Work `bson:"-"`
	GobWork  []byte

	RepoPath  string
	Workspace bool

	Link       string `bson:",omitempty"`
	Name       string `bson:",omitempty"`
	ImportPath string `bson:",omitempty"`
	Blurb      string `bson:",omitempty"`
}

func (w *Work) Freeze() {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(&w.Work); err != nil {
		panic(err)
	}
	w.GobWork = buf.Bytes()
}

func (w *Work) Thaw() {
	r := bytes.NewReader(w.GobWork)
	dec := gob.NewDecoder(r)
	if err := dec.Decode(&w.Work); err != nil {
		panic(err)
	}
}

func (w *Work) WholeID() string {
	return w.ID
}

type Build struct {
	TaskInfo `bson:",inline"`
	WorkID   string
	Build    builder.Build `bson:"-"`
	GobBuild []byte

	poke chan bool
}

func (b *Build) Freeze() {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(&b.Build); err != nil {
		panic(err)
	}
	b.GobBuild = buf.Bytes()
}

func (b *Build) Thaw() {
	r := bytes.NewReader(b.GobBuild)
	dec := gob.NewDecoder(r)
	if err := dec.Decode(&b.Build); err != nil {
		panic(err)
	}
}

func (b *Build) cleanup(num int) {
	defer log.Println(b.WholeID(), "clean up")
	defer b.Build.Cleanup()

	for i := 0; i < num; i++ {
		_, ok := <-b.poke
		if !ok {
			return
		}
	}
}

func (b *Build) WholeID() string {
	return fmt.Sprintf("%s:%s", b.WorkID, b.ID)
}

type Test struct {
	TaskInfo `bson:",inline"`
	WorkID   string
	BuildID  string
	Path     string

	Output   string
	Passed   bool
	Started  time.Time
	Duration time.Duration

	done chan bool
}

func (t *Test) Start() {
	if t.Started.IsZero() {
		t.Started = time.Now()
	}
}

func (t *Test) Finish() {
	t.Duration = time.Since(t.Started)
	t.done <- true
}

func (t *Test) WholeID() string {
	return fmt.Sprintf("%s:%s:%s", t.WorkID, t.BuildID, t.ID)
}

func new_info() (t TaskInfo) {
	t.When = time.Now()
	t.ID = new_id()
	return
}

func new_test(path string, build *Build, work *Work) (t *Test) {
	t = &Test{
		TaskInfo: new_info(),
		Path:     path,
		BuildID:  build.ID,
		WorkID:   work.ID,

		done: build.poke,
	}

	return
}

func new_build(build builder.Build, work *Work) (b *Build) {
	b = &Build{
		TaskInfo: new_info(),
		Build:    build,
		WorkID:   work.ID,

		poke: make(chan bool),
	}
	return
}

func new_work(work builder.Work) (w *Work) {
	type Linker interface {
		Link() string
	}
	type Namer interface {
		ProjectName() string
	}
	type Blurber interface {
		Blurb() string
	}

	w = &Work{
		TaskInfo:   new_info(),
		Work:       work,
		ImportPath: work.ImportPath(),
		RepoPath:   work.RepoPath(),
		Workspace:  work.IsWorkspace(),
	}
	if l, ok := work.(Linker); ok {
		w.Link = l.Link()
	}
	if n, ok := work.(Namer); ok {
		w.Name = n.ProjectName()
	}
	if b, ok := work.(Blurber); ok {
		w.Blurb = b.Blurb()
	}

	return
}
