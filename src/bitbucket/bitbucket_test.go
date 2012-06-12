package bitbucket

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
		t.Errorf("Expeced %+v. Got %+v", u.expect, ex)
	}
}

type testCase interface {
	perform(t *testing.T)
}

const example_packet = `{
  "repository": {
    "website": "http://foo.bar/",
    "fork": false,
    "name": "broker-test",
    "scm": "hg",
    "absolute_url": "/zeebo/broker-test/",
    "owner": "zeebo",
    "slug": "broker-test",
    "is_private": true
  },
  "commits": [    {
      "node": "5f46ab03da45",
      "files": [        {
          "type": "added",
          "file": "bitbucket"
        }],
      "branch": "default",
      "utctimestamp": "2012-06-12 14:31:15+00:00",
      "author": "zeebo",
      "timestamp": "2012-06-12 16:31:15",
      "raw_node": "5f46ab03da45361d8d3b5a35a8f6a54298934c70",
      "parents": [],
      "raw_author": "zeebo ",
      "message": "added bitbucket",
      "size": -1,
      "revision": 0
    },    {
      "node": "76eb9439c0e7",
      "files": [        {
          "type": "removed",
          "file": "bitbucket"
        }],
      "branch": "default",
      "utctimestamp": "2012-06-12 14:31:21+00:00",
      "author": "zeebo",
      "timestamp": "2012-06-12 16:31:21",
      "raw_node": "76eb9439c0e7a95c0b47a8036faf25b2bcd54e4c",
      "parents": [        "5f46ab03da45"],
      "raw_author": "zeebo ",
      "message": "removed bitbucket",
      "size": -1,
      "revision": 1
    }],
  "canon_url": "https://bitbucket.org",
  "user": "zeebo"
}`

func TestUnserialize(t *testing.T) {
	cases := []unserializeTest{
		{[]byte(example_packet), HookMessage{
			Repository: Repository{
				Website:      "http://foo.bar/",
				Fork:         false,
				Name:         "broker-test",
				SCM:          "hg",
				Absolute_URL: "/zeebo/broker-test/",
				Owner:        "zeebo",
				Slug:         "broker-test",
				Is_Private:   true,
			},
			Commits: []Commit{
				{
					Node: "5f46ab03da45",
					Files: []File{
						{Type: "added", File: "bitbucket"},
					},
					Author:   "zeebo",
					Raw_Node: "5f46ab03da45361d8d3b5a35a8f6a54298934c70",
					Message:  "added bitbucket",
				},
				Commit{
					Node: "76eb9439c0e7",
					Files: []File{
						{Type: "removed", File: "bitbucket"},
					},
					Author:   "zeebo",
					Raw_Node: "76eb9439c0e7a95c0b47a8036faf25b2bcd54e4c",
					Message:  "removed bitbucket",
				},
			},
			User:      "zeebo",
			Canon_URL: "https://bitbucket.org",
			Workspace: false,
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
	if v, ex := ex.RepoPath(), "https://bitbucket.org/zeebo/broker-test"; v != ex {
		t.Fatalf("Expected %+v. Got %+v", ex, v)
	}
	if v, ex := ex.Revisions(), []string{
		"5f46ab03da45361d8d3b5a35a8f6a54298934c70",
		"76eb9439c0e7a95c0b47a8036faf25b2bcd54e4c",
	}; !reflect.DeepEqual(v, ex) {
		t.Fatalf("Expected %+v. Got %+v", ex, v)
	}
}
