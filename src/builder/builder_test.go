package builder

import "testing"

type testWork struct {
	revisions  []string
	vcs        VCS
	repoPath   string
	importPath string
	workspace  bool
}

func (t *testWork) Revisions() []string { return t.revisions }
func (t *testWork) VCS() VCS            { return t.vcs }
func (t *testWork) RepoPath() string    { return t.repoPath }
func (t *testWork) ImportPath() string  { return t.importPath }
func (t *testWork) IsWorkspace() bool   { return t.workspace }

var _ Work = &testWork{}

func TestRun(t *testing.T) {
	w := &testWork{
		revisions: []string{
			"e4ef402bacb2a4e0a86c0729ffd531e52eb68d52",
			"34aa918aab43351e5ee86180cb170dc5b68f7a56",
		},
		vcs:        Git,
		repoPath:   "git://github.com/zeebo/irc",
		importPath: "github.com/zeebo/irc",
		workspace:  false,
	}

	reps, err := Run(w)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v", reps)
}
