package builder

import (
	"strings"
	"testing"
)

type existsWorld struct {
	t *testing.T
}

func (existsWorld) Exists(string) bool                    { return true }
func (existsWorld) LookPath(i string) (string, error)     { return i, nil }
func (existsWorld) TempDir(prefix string) (string, error) { return "/tmp/" + prefix, nil }

func (e existsWorld) Make(c command) (p proc) {
	return runner{e.t, c}
}

type runner struct {
	t *testing.T
	c command
}

func (r runner) Run() (error, bool) {
	t, c := r.t, r.c

	t.Log(c.args)

	//hack for expected output on getting timestamp
	if c.args[1] == "parents" && c.args[3] == "{date|rfc822date}" {
		c.w.Write([]byte(vcsHg.Format))
	}

	if c.args[1] == "log" {
		c.w.Write([]byte(vcsGit.Format))
	}

	//hack for expected output on finding revision
	if c.args[1] == "rev-parse" {
		c.w.Write([]byte("revision"))
	}
	if c.args[1] == "parents" && c.args[3] != "{date|rfc822date}" {
		c.w.Write([]byte("revision"))
	}

	//just print out the package it is listing. if it ends with a
	//... just print it out with a crazy suffix
	if c.args[1] == "list" {
		pack := c.args[len(c.args)-1]
		if strings.HasSuffix(pack, "...") {
			pack = pack[:len(pack)-3]
			c.w.Write([]byte(pack))
			c.w.Write([]byte("foo\n"))
		}
		c.w.Write([]byte(pack))
	}

	return nil, true
}

func TestMocked(t *testing.T) {
	//mock out the world
	defer func(e environ) { world = e }(world)
	world = existsWorld{t}

	works := []*Work{
		{
			Revision:   "e9dd26552f10d390b5f9f59c6a9cfdc30ed1431c",
			ImportPath: "github.com/zeebo/irc",
		},
		{
			Revision:    "cc5e03949586c5b697c9a3080cc3bf7501a14d96",
			ImportPath:  "github.com/dustin/githubhooks",
			Subpackages: true,
		},
		{
			ImportPath: "github.com/zeebo/irc",
		},
	}

	for _, w := range works {
		_, _, err := New("", "", "", "").Build(w)
		if err != nil {
			t.Error(err)
		} else {
			t.Log("========")
			t.Logf("%s passed", w.ImportPath)
		}
	}
}
