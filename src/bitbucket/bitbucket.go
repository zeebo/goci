package bitbucket

import (
	"builder"
	"encoding/gob"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"path"
)

func init() {
	gob.Register(&HookMessage{})
}

type HookMessage struct {
	Repository Repository
	Commits    []Commit
	User       string
	Canon_URL  string

	Workspace bool
}

type Repository struct {
	Name         string
	SCM          string
	Absolute_URL string
	Owner        string
	Slug         string
	Is_Private   bool
}

type Commit struct {
	Node     string
	Files    []File
	Author   string
	Raw_Node string
	Message  string
}

type File struct {
	Type, File string
}

var UnknownVCS = errors.New("Unknown VCS")

func (h *HookMessage) checkVcs() (err error) {
	switch h.Repository.SCM {
	case "git", "hg":
	default:
		err = UnknownVCS
	}
	return
}

func (h *HookMessage) Load(r io.Reader) (err error) {
	dec := json.NewDecoder(r)
	err = dec.Decode(h)
	if err != nil {
		return
	}
	err = h.checkVcs()
	return
}

func (h *HookMessage) LoadBytes(p []byte) (err error) {
	err = json.Unmarshal(p, h)
	if err != nil {
		return
	}
	err = h.checkVcs()
	return
}

func (h *HookMessage) RepoPath() (p string) {
	parsed := url.URL{
		Host:   "bitbucket.org",
		Path:   path.Clean(h.Repository.Absolute_URL),
		Scheme: "https",
	}
	if h.Repository.SCM == "git" {
		parsed.Scheme = "git"
		parsed.Path += ".git"
	}

	p = parsed.String()
	return
}

func (h *HookMessage) ImportPath() (p string) {
	p = path.Clean(path.Join("bitbucket.org", h.Repository.Absolute_URL))
	return
}

func (h *HookMessage) VCS() builder.VCS {
	switch h.Repository.SCM {
	case "git":
		return builder.Git
	case "hg":
		return builder.HG
	}
	return nil
}

func (h *HookMessage) WorkType() builder.WorkType {
	if h.Workspace {
		return builder.WorkTypeWorkspace
	}
	return builder.WorkTypePackage
}

func (h *HookMessage) Revisions() (revs []string) {
	for _, ci := range h.Commits {
		revs = append(revs, ci.Raw_Node)
	}
	return
}

func (h *HookMessage) ProjectName() string {
	return h.Repository.Name
}

func (h *HookMessage) Link() string {
	return h.Repository.Website
}

//ensure our HookMessage is a valid work item
var _ builder.Work = &HookMessage{}
