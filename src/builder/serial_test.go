package builder

import "testing"

const test_work_json = `{"revisions":["e4ef402bacb2a4e0a86c0729ffd531e52eb68` +
	`d52","34aa918aab43351e5ee86180cb170dc5b68f7a56"],"vc` +
	`s":"git","repopath":"git://github.com/zeebo/irc","im` +
	`portpath":"github.com/zeebo/irc","workspace":false}`

func TestSerialize(t *testing.T) {
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

	o, err := Serialize(w)
	if err != nil {
		t.Fatal(err)
	}
	if o != test_work_json {
		t.Errorf("Expected: %q\nGot: %q", test_work_json, o)
	}
}

func TestUnserialize(t *testing.T) {
	w, err := Unserialize(test_work_json)
	if err != nil {
		t.Fatal(err)
	}

	if g, ex := w.VCS(), Git; g != ex {
		t.Errorf("Expected %q\nGot %q", ex, g)
	}

	if g, ex := w.RepoPath(), "git://github.com/zeebo/irc"; g != ex {
		t.Errorf("Expected %q\nGot %q", ex, g)
	}

	if g, ex := w.ImportPath(), "github.com/zeebo/irc"; g != ex {
		t.Errorf("Expected %q\nGot %q", ex, g)
	}

	if g, ex := w.IsWorkspace(), false; g != ex {
		t.Errorf("Expected %q\nGot %q", ex, g)
	}
}
