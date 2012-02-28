package main

import "testing"

func TestRepoHash(t *testing.T) {
	cases := []struct {
		r  Repo
		ex string
	}{
		{"foo", "0beec7b5ea3f0fdbc95d0dd47f3c5bc275da8a33"},
		{"bar", "62cdb7020ff920e5aa642c3d4066950dd1f01f4d"},
		{"baz", "bbe960a25ea311d21d40669e93df2003ba9b90a2"},
		{"http://github.com/zeebo/goci", "f801fa31e64e00d31e65eb52beeb957f3667cc5c"},
	}

	for _, test := range cases {
		if ex := test.r.Hash(); ex != test.ex {
			t.Errorf("%q: Expected %q Got %q", test.r, test.ex, ex)
		}
	}
}
