package builder

import "testing"

func TestRunTool(t *testing.T) {
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

	for _, r := range reps {
		t.Logf("%q", r)
	}
}

func TestRunWorkspace(t *testing.T) {
	w := &testWork{
		revisions: []string{
			"6d1ed8f9512102f30227ebfe8a327a572cbae7f2",
			"48d02e161b71b9ccce7d4e91439f43628f215003",
		},
		vcs:        Git,
		repoPath:   "git://github.com/goods/starter",
		importPath: "",
		workspace:  true,
	}

	reps, err := Run(w)
	if err != nil {
		t.Fatal(err)
	}

	for _, r := range reps {
		t.Logf("%q", r)
	}
}
