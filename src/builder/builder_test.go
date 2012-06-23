// +build !goci

package builder

import "testing"

func TestGoGetPackage(t *testing.T) {
	w := &testWork{
		importPath: "github.com/zeebo/irc",
		workType:   WorkTypeGoinstall,
	}

	reps, err := CreateBuilds(w)
	if err != nil {
		t.Fatal(err)
	}

	for _, r := range reps {
		t.Logf("%q", r)
		t.Log("cleanup:", r.Cleanup())
	}
}

func TestGoGetPackageTestDeps(t *testing.T) {
	w := &testWork{
		importPath: "labix.org/v2/mgo",
		workType:   WorkTypeGoinstall,
	}

	reps, err := CreateBuilds(w)
	if err != nil {
		t.Fatal(err)
	}

	for _, r := range reps {
		t.Logf("%q", r)
		t.Log("cleanup:", r.Cleanup())
	}
}

func TestCreateBuildsNonWorkspace(t *testing.T) {
	w := &testWork{
		revisions: []string{
			"e4ef402bacb2a4e0a86c0729ffd531e52eb68d52", //empty tests
			"1351a526989eda49cf7159561f38d9454c8e961a", //before go1 commit
			"c97d0b46f86c1d1294b9351c01349177e38ef2b3", //working with tests
			"b1a6b6797e2009e1dac7ccd5515f8aee17df6774", //tests that fail to compile
		},
		vcs:        Git,
		repoPath:   "git://github.com/zeebo/irc",
		importPath: "github.com/zeebo/irc",
		workType:   WorkTypePackage,
	}

	reps, err := CreateBuilds(w)
	if err != nil {
		t.Fatal(err)
	}

	for _, r := range reps {
		t.Logf("%q", r)
		t.Log("cleanup:", r.Cleanup())
	}
}

func TestExternalThings(t *testing.T) {
	w := &testWork{
		revisions: []string{
			"467d3ae22642ddadbaf6a0693c02d18b24fb7d35",
			"467d3ae22642ddadbaf6a0693c02d18b24fb7d35", //test it twice
		},
		vcs:        Git,
		repoPath:   "git://github.com/ftrvxmtrx/omgfsm",
		importPath: "github.com/ftrvxmtrx/omgfsm",
		workType:   WorkTypePackage,
	}

	reps, err := CreateBuilds(w)
	if err != nil {
		t.Fatal(err)
	}

	for _, r := range reps {
		t.Logf("%q", r)
		t.Log("cleanup:", r.Cleanup())
	}
}

func TestStrangeLayout(t *testing.T) {
	w := &testWork{
		revisions: []string{
			"8597510c796a7214804c507dcc879dcd474d547c",
		},
		vcs:        HG,
		repoPath:   "https://code.google.com/p/go-charset",
		importPath: "code.google.com/p/go-charset",
		workType:   WorkTypePackage,
	}

	reps, err := CreateBuilds(w)
	if err != nil {
		t.Fatal(err)
	}

	for _, r := range reps {
		t.Logf("%q", r)
		t.Log("cleanup:", r.Cleanup())
	}
}

func TestCreateBuildsWorkspace(t *testing.T) {
	w := &testWork{
		revisions: []string{
			"6d1ed8f9512102f30227ebfe8a327a572cbae7f2",
			"48d02e161b71b9ccce7d4e91439f43628f215003",
		},
		vcs:        Git,
		repoPath:   "git://github.com/goods/starter",
		importPath: "",
		workType:   WorkTypeWorkspace,
	}

	reps, err := CreateBuilds(w)
	if err != nil {
		t.Fatal(err)
	}

	for _, r := range reps {
		t.Logf("%q", r)
		t.Log("cleanup:", r.Cleanup())
	}
}

func TestCreateBuildsWorkspaceGoci(t *testing.T) {
	w := &testWork{
		revisions: []string{
			"185f3c1735c482a280d00ae09f15b4f6b05f6d22",
		},
		vcs:        Git,
		repoPath:   "git://github.com/zeebo/goci",
		importPath: "",
		workType:   WorkTypeWorkspace,
	}

	reps, err := CreateBuilds(w)
	if err != nil {
		t.Fatal(err)
	}

	for _, r := range reps {
		t.Logf("%q", r)
		t.Log("cleanup:", r.Cleanup())
	}
}
