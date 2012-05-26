package builder

import (
	"encoding/json"
	"fmt"
)

type work struct {
	Revs      []string `json:"revisions"`
	Vcs       string   `json:"vcs"`
	Repo      string   `json:"repopath"`
	Import    string   `json:"importpath"`
	Workspace bool     `json:"workspace"`
}

func (t *work) Revisions() []string { return t.Revs }
func (t *work) RepoPath() string    { return t.Repo }
func (t *work) ImportPath() string  { return t.Import }
func (t *work) IsWorkspace() bool   { return t.Workspace }

func (t *work) VCS() (v VCS) {
	switch t.Vcs {
	case "git":
		v = Git
	case "hg":
		v = HG
	}
	return
}

func Serialize(w Work) (out string, err error) {
	t := work{
		Revs:      w.Revisions(),
		Repo:      w.RepoPath(),
		Import:    w.ImportPath(),
		Workspace: w.IsWorkspace(),
	}
	switch w.VCS() {
	case Git:
		t.Vcs = "git"
	case HG:
		t.Vcs = "hg"
	default:
		err = fmt.Errorf("Unsupported vcs")
		return
	}

	ob, err := json.Marshal(t)
	if err != nil {
		return
	}
	out = string(ob)
	return
}

func Unserialize(in string) (w Work, err error) {
	var t work
	err = json.Unmarshal([]byte(in), &t)
	if err != nil {
		return
	}
	w = &t
	return
}
