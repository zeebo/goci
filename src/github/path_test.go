package github

import "testing"

func TestParsePath(t *testing.T) {
	cases := []struct {
		path, ex string
	}{
		{"foo", "bar"},
		{"foo", "bar"},
		{"foo", "bar"},
	}

	for _, test := range cases {
		ex, err := ClonePath(test.path)
		if err != nil {
			t.Errorf("%s: %s", test.path, err)
		}
		if ex != test.ex {
			t.Errorf("Expected %q. Got %q", test.ex, ex)
		}
	}
}
