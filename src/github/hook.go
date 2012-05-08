package github

import (
	"builder"
	"encoding/json"
	"io"
	"net/url"
	"time"
)

type HookMessage struct {
	Before, After, Ref string
	Commits            []Commit
	Repository         Repository
	Pusher             Author
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
	ID, Message              string
	Timestamp                time.Time
	URL                      string
	Added, Removed, Modified []string
	Author                   Author
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

func (h *HookMessage) ClonePath() (path string) {
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

func (h *HookMessage) Revisions() (revs []string) {
	return
}

//ensure our HookMessage is a valid work item
var (
	_ builder.ToolWork   = &HookMessage{}
	_ builder.GopathWork = &HookMessage{}
)
