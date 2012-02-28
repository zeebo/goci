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
		{"git://github.com/zeebo/goci.git", "0e70ab7208fd12fde38bb19bd171a5e39d48751e"},
	}

	for _, test := range cases {
		if ex := test.r.Hash(); ex != test.ex {
			t.Errorf("%q: Expected %q Got %q", test.r, test.ex, ex)
		}
	}
}

func TestCloneAndTest(t *testing.T) {
	repo := Repo("git://github.com/zeebo/heroku-basic-app.git")
	defer repo.Cleanup()
	if err := repo.Clone(); err != nil {
		t.Fatal("clone:", err)
	}

	//list
	packages, err := repo.Packages()
	if err != nil {
		t.Fatal("list:", err)
	}

	//build
	stdout, stderr, err := repo.Get()
	if err != nil {
		t.Error("get:", err)
		t.Log(stdout.String())
		t.Log(stderr.String())
		t.FailNow()
	}
	t.Log(stdout.String())

	//test
	stdout, stderr, err = repo.Test(packages)
	if err != nil {
		t.Error("test:", err)
		t.Log(stdout.String())
		t.Log(stderr.String())
		t.FailNow()
	}
	t.Log(stdout.String())
}
