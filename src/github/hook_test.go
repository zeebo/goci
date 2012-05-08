package github

import (
	"reflect"
	"testing"
	"time"
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
"before": "5aef35982fb2d34e9d9d4502f6ede1072793222d",
  "repository": {
    "url": "http://github.com/defunkt/github",
    "name": "github",
    "description": "You're lookin' at it.",
    "watchers": 5,
    "forks": 2,
    "private": true,
    "owner": {
      "email": "chris@ozmm.org",
      "name": "defunkt"
    }
  },
  "commits": [
    {
      "id": "41a212ee83ca127e3c8cf465891ab7216a705f59",
      "url": "http://github.com/defunkt/github/commit/41a212ee83ca127e3c8cf465891ab7216a705f59",
      "author": {
        "email": "chris@ozmm.org",
        "name": "Chris Wanstrath"
      },
      "message": "okay i give in",
      "timestamp": "2008-02-15T14:57:17-08:00",
      "added": ["filepath.rb"]
    },
    {
      "id": "de8251ff97ee194a289832576287d6f8ad74e3d0",
      "url": "http://github.com/defunkt/github/commit/de8251ff97ee194a289832576287d6f8ad74e3d0",
      "author": {
        "email": "chris@ozmm.org",
        "name": "Chris Wanstrath"
      },
      "message": "update pricing a tad",
      "timestamp": "2008-02-15T14:36:34-08:00"
    }
  ],
  "after": "de8251ff97ee194a289832576287d6f8ad74e3d0",
  "ref": "refs/heads/master"
}`

func TestUnserialize(t *testing.T) {
	cases := []unserializeTest{
		{[]byte(example_packet), HookMessage{
			Before: "5aef35982fb2d34e9d9d4502f6ede1072793222d",
			Repository: Repository{
				URL:         "http://github.com/defunkt/github",
				Name:        "github",
				Description: "You're lookin' at it.",
				Watchers:    5,
				Forks:       2,
				Private:     true,
				Owner: Author{
					Email: "chris@ozmm.org",
					Name:  "defunkt",
				},
			},
			Commits: []Commit{
				{
					ID:  "41a212ee83ca127e3c8cf465891ab7216a705f59",
					URL: "http://github.com/defunkt/github/commit/41a212ee83ca127e3c8cf465891ab7216a705f59",
					Author: Author{
						Email: "chris@ozmm.org",
						Name:  "Chris Wanstrath",
					},
					Message:   "okay i give in",
					Timestamp: time.Date(2008, 02, 15, 14, 57, 17, 0, time.FixedZone("", -8*3600)),
					Added:     []string{"filepath.rb"},
				},
				{
					ID:  "de8251ff97ee194a289832576287d6f8ad74e3d0",
					URL: "http://github.com/defunkt/github/commit/de8251ff97ee194a289832576287d6f8ad74e3d0",
					Author: Author{
						Email: "chris@ozmm.org",
						Name:  "Chris Wanstrath",
					},
					Message:   "update pricing a tad",
					Timestamp: time.Date(2008, 02, 15, 14, 36, 34, 0, time.FixedZone("", -8*3600)),
				},
			},
			After: "de8251ff97ee194a289832576287d6f8ad74e3d0",
			Ref:   "refs/heads/master",
		}},
	}

	for _, c := range cases {
		c.perform(t)
	}
}

func TestHookMessagePaths(t *testing.T) {
	var ex HookMessage
	if err := ex.LoadBytes([]byte(example_packet)); err != nil {
		t.Fatal(err)
	}
	if v, ex := ex.ClonePath(), "git://github.com/defunkt/github.git"; v != ex {
		t.Fatalf("Expected %+v. Got %+v", ex, v)
	}
	if v, ex := ex.ImportPath(), "github.com/defunkt/github"; v != ex {
		t.Fatalf("Expected %+v. Got %+v", ex, v)
	}
	if v, ex := ex.Revisions(), []string{
		"41a212ee83ca127e3c8cf465891ab7216a705f59",
		"de8251ff97ee194a289832576287d6f8ad74e3d0",
	}; !reflect.DeepEqual(v, ex) {
		t.Fatalf("Expected %+v. Got %+v", ex, v)
	}
}
