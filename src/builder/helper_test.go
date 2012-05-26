package builder

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
