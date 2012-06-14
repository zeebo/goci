package google

import (
	"reflect"
	"testing"
)

type unserializeTest struct {
	payload []byte
	expect  HookMessage
}

func (u *unserializeTest) perform(t *testing.T) {
	var ex HookMessage
	if err := ex.LoadBytes(u.payload); err != nil {
		t.Errorf("Error unmarshaling payload: %s", err)
		return
	}
	if !reflect.DeepEqual(ex, u.expect) {
		t.Errorf("Expeced %#v. Got %#v", u.expect, ex)
	}
}

type testCase interface {
	perform(t *testing.T)
}

const example_packet = `{
	"project_name": "goci",
	"repository_path": "https://code.google.com/p/goci/",
	"revision_count": 1,
	"revisions": [
		{
			"added": [ "/src/google/google_test.go" ],
			"author": "jeff <leterip@me.com>",
			"message": "added stub test file for google messages\n",
			"modified": [],
			"path_count": 1,
			"removed": [],
			"revision": "49f4264c954ee50aef8cd6826a17a8d1f8500ab1",
			"timestamp": 1339708892,
			"url": "http://goci.googlecode.com/git-history/49f4264c954ee50aef8cd6826a17a8d1f8500ab1/"
		}
	]
}`

func TestUnserialize(t *testing.T) {
	cases := []unserializeTest{
		{[]byte(example_packet), HookMessage{
			Project_Name:    "goci",
			Repository_Path: "https://code.google.com/p/goci/",
			Revision_Count:  1,
			Commits: []Commit{
				{
					Revision:   "49f4264c954ee50aef8cd6826a17a8d1f8500ab1",
					URL:        "http://goci.googlecode.com/git-history/49f4264c954ee50aef8cd6826a17a8d1f8500ab1/",
					Author:     "jeff <leterip@me.com>",
					Timestamp:  1339708892,
					Message:    "added stub test file for google messages\n",
					Path_Count: 1,
					Added:      []string{"/src/google/google_test.go"},
					Modified:   []string{},
					Removed:    []string{},
				},
			},
			Workspace: false,
			Vcs:       nil,
		},
		}}

	for _, c := range cases {
		c.perform(t)
	}
}

func TestHookMessagePaths(t *testing.T) {
	var ex HookMessage
	if err := ex.LoadBytes([]byte(example_packet)); err != nil {
		t.Fatal(err)
	}
	if v, ex := ex.RepoPath(), "https://code.google.com/p/goci"; v != ex {
		t.Fatalf("Expected %+v. Got %+v", ex, v)
	}
	if v, ex := ex.Revisions(), []string{
		"49f4264c954ee50aef8cd6826a17a8d1f8500ab1",
	}; !reflect.DeepEqual(v, ex) {
		t.Fatalf("Expected %#v. Got %#v", ex, v)
	}
}
