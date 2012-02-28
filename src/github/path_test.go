package github

import "testing"

func TestGetRepo(t *testing.T) {
	cases := []struct {
		path, repo string
	}{
		{"http://github.com/zeebo/goci", "zeebo/goci"},
		{"http://github.com/user/repo", "user/repo"},
		{"http://github.com/blah/boof", "blah/boof"},
		{"http://goofub.com/nope/not", ""},
	}

	for _, test := range cases {
		if repo := getRepo(test.path); repo != test.repo {
			t.Errorf("Expected %q. Got %q", test.repo, repo)
		}
	}
}

func TestClonePath(t *testing.T) {
	cases := []struct {
		path, repo string
	}{
		{"http://github.com/zeebo/goci", "git://github.com/zeebo/goci.git"},
		{"http://github.com/user/repo", "git://github.com/user/repo.git"},
		{"http://github.com/blah/boof", "git://github.com/blah/boof.git"},
	}

	for _, test := range cases {
		repo, err := ClonePath(test.path)
		if err != nil {
			t.Errorf("%q: %s", test.path, err)
		}
		if repo != test.repo {
			t.Errorf("%q: Expected %q. Got %q", test.path, test.repo, repo)
		}
	}
}

func TestInvalidClonePath(t *testing.T) {
	cases := []string{
		"http://githob.com/zeebo/goci",
	}

	for _, test := range cases {
		_, err := ClonePath(test)
		if err == nil {
			t.Errorf("%q: expected error.", test)
		}
	}
}
