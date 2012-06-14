package google

import (
	"encoding/gob"
	"encoding/json"
	"builder"
	"errors"
	"io"
	"net/url"
	"path"
)

func init() {
	gob.Register(&HookMessage{})
}

type HookMessage struct {
	Project_Name    string
	Repository_Path string
	Revision_Count  int
	Commits         []Commit `json:"revisions"`

	Workspace bool
	Vcs       builder.VCS
}

type Commit struct {
	Revision   Rev
	URL        string
	Author     string
	Timestamp  int64
	Message    string
	Path_Count string

	Added    []string
	Modified []string
	Removed  []string
}

type Rev string

func (r *Rev) UnmarshalJSON(p []byte) (err error) {
	if len(p) == 0 {
		err = errors.New("empty revision")
		return
	}

	switch p[0] {
	case '"':
		if len(p) == 1 {
			err = errors.New("unended quote on revision")
			return
		}
		*r = Rev(p[1 : len(p)-1])
	default:
		*r = Rev(p)
	}
	return
}

func (r Rev) String() string {
	return string(r)
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

func (h *HookMessage) RepoPath() (p string) {
	parsed := url.URL{
		Host:   "code.google.com",
		Path:   path.Join("p", h.Project_Name),
		Scheme: "https",
	}
	//TODO: support svn
	//http://goci.googlecode.com/svn/ goci-read-only
	p = parsed.String()
	return
}

func (h *HookMessage) ImportPath() (p string) {
	p = path.Join("code.google.com", "p", h.Project_Name)
	return
}

func (h *HookMessage) VCS() builder.VCS {
	return h.Vcs
}

func (h *HookMessage) IsWorkspace() bool {
	return h.Workspace
}

func (h *HookMessage) Revisions() (revs []string) {
	for _, ci := range h.Commits {
		revs = append(revs, ci.Revision.String())
	}
	return
}

func (h *HookMessage) ProjectName() string {
	return h.Project_Name
}

func (h *HookMessage) Link() string {
	return h.Repository_Path
}
