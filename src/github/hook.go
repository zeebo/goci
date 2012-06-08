package github

import (
	"builder"
	"encoding/gob"
	"encoding/json"
	"io"
	"net/url"
	"time"
)

func init() {
	gob.Register(&HookMessage{})
}

type HookMessage struct {
	Before, After, Ref string

	Commits    []Commit
	Repository Repository
	Pusher     Author
	Workspace  bool
}

type Repository struct {
	Name                string
	URL, Homepage       string
	Pledie, Description string
	Watchers, Forks     int
	Private             bool
	Owner               Author
}

type Author struct {
	Name, Email string
}

type Commit struct {
	ID, Message string
	Timestamp   time.Time
	URL         string
	Author      Author

	Added, Removed, Modified []string
}

func (h *HookMessage) Load(r io.Reader) (err error) {
	dec := json.NewDecoder(r)
	err = dec.Decode(h)
	return
}

func (h *HookMessage) LoadBytes(p []byte) (err error) {
	err = json.Unmarshal(p, h)
	return
}

func (h *HookMessage) RepoPath() (path string) {
	parsed, err := url.Parse(h.Repository.URL)
	if err != nil {
		panic(err)
	}
	parsed.Scheme = "git"
	parsed.Path += ".git"

	path = parsed.String()
	return
}

func (h *HookMessage) ImportPath() (path string) {
	parsed, err := url.Parse(h.Repository.URL)
	if err != nil {
		panic(err)
	}
	path = parsed.Host + parsed.Path
	return
}

func (h *HookMessage) VCS() builder.VCS {
	return builder.Git
}

func (h *HookMessage) IsWorkspace() bool {
	return h.Workspace
}

func (h *HookMessage) Revisions() (revs []string) {
	for _, ci := range h.Commits {
		revs = append(revs, ci.ID)
	}
	return
}

func (h *HookMessage) ProjectName() string {
	return h.Repository.Name
}

func (h *HookMessage) Link() string {
	return h.Repository.URL
}

func (h *HookMessage) Blurb() string {
	return h.Repository.Description
}

//ensure our HookMessage is a valid work item
var _ builder.Work = &HookMessage{}
