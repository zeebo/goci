package worker

import (
	"builder"
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"time"
	fp "path/filepath"
)

type Bytes []byte

func (b Bytes) Format(s fmt.State, c rune) {
	fmt.Fprint(s, "[]byte{...}")
}

type WorkStatus int

const (
	WorkStatusPassed = iota
	WorkStatusFailed
	WorkStatusWary
)

func (w WorkStatus) String() (r string) {
	switch w {
	case WorkStatusPassed:
		r = "Passed"
	case WorkStatusFailed:
		r = "Failed"
	case WorkStatusWary:
		r = "Be Wary"
	default:
		panic("unknown work status")
	}
	return
}

func (w WorkStatus) LabelType() (r string) {
	switch w {
	case WorkStatusPassed:
		r = "success"
	case WorkStatusFailed:
		r = "important"
	case WorkStatusWary:
		r = "warning"
	default:
		panic("unknown work status")
	}
	return
}

func (w WorkStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, w)), nil
}

type Work struct {
	When  time.Time
	ID    string `bson:"_id" json:"-"`
	Error string `json:",omitempty"`

	Work    builder.Work `bson:"-" json:"-"`
	GobWork Bytes        `json:"-"`
	Builds  []*Build

	RepoPath string
	WorkType builder.WorkType
	Status   WorkStatus

	Link       string `bson:",omitempty" json:",omitempty"`
	Name       string `bson:",omitempty" json:",omitempty"`
	ImportPath string `bson:",omitempty" json:",omitempty"`
	Blurb      string `bson:",omitempty" json:",omitempty"`

	poke chan *Build
}

type Build struct {
	ID    string `bson:"_id" json:"-"`
	Error string `json:",omitempty"`

	WorkID   string        `json:"-"`
	Build    builder.Build `bson:"-" json:"-"`
	GobBuild Bytes         `json:"-"`
	Tests    []*Test

	Revision string
	Passed   bool

	poke chan *Test
	done chan *Build //ref to the work channel
}

type Test struct {
	ID    string `bson:"_id" json:"-"`
	Error string `json:",omitempty"`

	WorkID  string `json:"-"`
	BuildID string `json:"-"`
	Path    string `json:"-"`

	Output   string `json:",omitempty"`
	Passed   bool
	Started  time.Time
	Duration time.Duration

	done chan *Test //ref to the build channel
}

func new_test(path string, build *Build, work *Work) (t *Test) {
	t = &Test{
		ID:      new_id(),
		Path:    path,
		BuildID: build.ID,
		WorkID:  work.ID,

		done: build.poke,
	}

	return
}

func new_build(build builder.Build, work *Work) (b *Build) {
	b = &Build{
		ID:       new_id(),
		Build:    build,
		WorkID:   work.ID,
		Revision: build.Revision(),
		Passed:   true,

		poke: make(chan *Test),
		done: work.poke,
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
		When:       time.Now(),
		ID:         new_id(),
		Work:       work,
		ImportPath: work.ImportPath(),
		RepoPath:   work.RepoPath(),
		WorkType:   work.WorkType(),

		poke: make(chan *Build),
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

func (w *Work) DisplayName() (r string) {
	switch {
	case w.ImportPath != "":
		r = w.ImportPath
	case w.Name != "":
		r = w.Name
	case w.RepoPath != "":
		r = w.RepoPath
	}
	return
}

func (w *Work) Freeze() {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(&w.Work); err != nil {
		panic(err)
	}
	w.GobWork = buf.Bytes()

	for _, b := range w.Builds {
		b.Freeze()
	}
}

func (w *Work) Thaw() {
	r := bytes.NewReader(w.GobWork)
	dec := gob.NewDecoder(r)
	if err := dec.Decode(&w.Work); err != nil {
		panic(err)
	}

	for _, b := range w.Builds {
		b.Thaw()
	}
}

func (w *Work) WholeID() string {
	return w.ID
}

func (w *Work) update_status() {
	var passed, failed = true, true
	for _, b := range w.Builds {
		passed = passed && b.Passed
		failed = failed && !b.Passed
	}
	switch {
	case passed && !failed:
		w.Status = WorkStatusPassed
	case failed && !passed:
		w.Status = WorkStatusFailed
	default:
		w.Status = WorkStatusWary
	}
}

func (w *Work) wait_for(num int, done chan bool) {
	defer func() {
		done <- true
	}()

	for i := 0; i < num; i++ {
		b, ok := <-w.poke
		if !ok {
			return
		}
		w.Builds = append(w.Builds, b)
	}
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
	defer func() { b.done <- b }()
	defer log.Println(b.WholeID(), "clean up")
	defer b.Build.Cleanup()

	for i := 0; i < num; i++ {
		t, ok := <-b.poke
		if !ok {
			return
		}
		b.Tests = append(b.Tests, t)
		b.Passed = b.Passed && t.Passed
	}
}

func (b *Build) WholeID() string {
	return fmt.Sprintf("%s:%s", b.WorkID, b.ID)
}

func (t *Test) Start() {
	if t.Started.IsZero() {
		t.Started = time.Now()
	}
}

func (t *Test) Finish() {
	t.Duration = time.Since(t.Started)
	t.done <- t
}

func (t *Test) WholeID() string {
	return fmt.Sprintf("%s:%s:%s", t.WorkID, t.BuildID, t.ID)
}

func (t *Test) BaseName() string {
	b := fp.Base(t.Path)
	ext := fp.Ext(b)
	return b[:len(b)-len(ext)]
}
