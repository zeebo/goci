package builder

import (
	"reflect"
	"testing"
)

func TestUniqueStrings(t *testing.T) {
	data := []struct {
		in, out []string
	}{
		{[]string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{[]string{"a", "b", "b", "c"}, []string{"a", "b", "c"}},
		{[]string{}, []string{}},
		{nil, nil},
		{[]string{"b", "c", "b", "d"}, []string{"b", "c", "d"}},
	}

	for i, v := range data {
		//do an in place unique
		t.Logf("%d: testing %v", i, v.in)
		out := unique(v.in)
		if !reflect.DeepEqual(out, v.out) {
			t.Errorf("%d: %+v != %+v", i, out, v.out)
		} else {
			t.Logf("%d: %+v == %+v", i, out, v.out)
		}
	}
}
