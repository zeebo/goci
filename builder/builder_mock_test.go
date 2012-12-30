package builder

import (
	"github.com/zeebo/goci/app/rpc"
	"github.com/zeebo/goci/environ"
	"github.com/zeebo/goci/gotool"
	"github.com/zeebo/goci/tarball"
	"github.com/zeebo/goci/vcs"
	"strings"
	"testing"
	"time"
)

func testMode(r environ.TestRun) (environ.TestEnv, func()) {
	p1, p2, p3, p4 := World, tarball.World, vcs.World, gotool.World
	tw := environ.NewTest(3)
	tw.SetRun(r)
	World, tarball.World, vcs.World, gotool.World = tw, tw, tw, tw
	return tw, func() {
		World, tarball.World, vcs.World, gotool.World = p1, p2, p3, p4
	}
}

func testRun(c environ.Command) (error, bool) {
	//giant switch of doom to mock out all the different commands
	switch c.Args[0] {
	case "hg":
		switch {
		case c.Args[1] == "parents" && c.Args[3] == "{date|rfc822date}": //date
			c.W.Write([]byte(time.RFC822Z))
		case c.Args[1] == "parents" && c.Args[3] == "{node}":
			c.W.Write([]byte(`revision`))
		}
	case "git":
		switch {
		case c.Args[1] == "log": //date
			c.W.Write([]byte("Mon, 2 Jan 2006 15:04:05 -0700"))
		case c.Args[1] == "rev-parse":
			c.W.Write([]byte(`revision`))
		}
	case "bzr":
		switch {
		case c.Args[1] == "log":
			c.W.Write([]byte(`timestamp: `))
			c.W.Write([]byte("Mon 2006-01-02 15:04:05 -0700"))
		case c.Args[1] == "revision-info":
			c.W.Write([]byte(`120 foobar revision`))
		}
	case "go":
		switch {
		case c.Args[1] == "list":
			pack := c.Args[len(c.Args)-1]
			if strings.HasSuffix(pack, "...") {
				pack = pack[:len(pack)-3]
				c.W.Write([]byte(pack))
				c.W.Write([]byte("foo\n"))
			}
			c.W.Write([]byte(pack))
		}
	}

	//phew
	return nil, true
}

func TestMocked(t *testing.T) {
	tw, und := testMode(environ.TestRun(testRun))
	defer und()

	works := []*rpc.Work{
		{
			Revision:   "e9dd26552f10d390b5f9f59c6a9cfdc30ed1431c",
			ImportPath: "github.com/zeebo/irc",
			VCSHint:    "git",
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
		for _, vcs := range []string{"git", "hg", "bzr"} {
			tw.Reset()
			w.VCSHint = vcs
			_, _, err := New("", "", "goroot").Build(w)
			if err != nil {
				t.Error(err)
				tw.Dump(t)
			} else {
				t.Logf("%s[%s] passed", w.ImportPath, vcs)
			}
		}
	}
}
