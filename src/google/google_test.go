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
		t.Errorf("Expeced %+v. Got %+v", u.expect, ex)
	}
}

type testCase interface {
	perform(t *testing.T)
}

const example_packet = ``

func TestUnserialize(t *testing.T) {
	cases := []unserializeTest{
		{[]byte(example_packet), HookMessage{}},
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
