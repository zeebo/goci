package environ

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	fp "path/filepath"
	"testing"
	"time"
)

//
// Logger
//

type logger struct {
	t      *testing.T
	events []string
}

func (l *logger) Logf(format string, vals ...interface{}) {
	s := fmt.Sprintf(format, vals...)
	// l.t.Logf("%q", s)
	l.events = append(l.events, s)
}

func (l *logger) Reset() {
	l.events = l.events[:0]
}

//
// Files
//

type testIO struct {
	t    *logger
	name string
}

func (i testIO) Close() error {
	i.t.Logf("%s: Close()", i.name)
	return nil
}

func (i testIO) Read(p []byte) (int, error) {
	i.t.Logf("%s: Read(%d)", i.name, len(p))
	return len(p), nil
}

func (i testIO) Write(p []byte) (int, error) {
	i.t.Logf("%s: Write(%d)", i.name, len(p))
	return len(p), nil
}

//
// Stat response
//

type testInfo struct {
	t    *logger
	name string
	size int64
}

func (t testInfo) String() string {
	return fmt.Sprintf("%s:%d", t.name, t.getSize())
}

func (t testInfo) getSize() (s int64) {
	if !t.IsDir() {
		s = t.size
	}
	return
}

func (t testInfo) Size() (s int64) {
	s = t.getSize()
	t.t.Logf("%v: Size(): %d", t.name, s)
	return
}

func (t testInfo) Mode() (m os.FileMode) {
	t.t.Logf("%v: Mode(): dir:%v", t.name, t.IsDir())
	if t.IsDir() {
		m = os.ModeDir
	}
	return
}

func (t testInfo) IsDir() bool {
	base := fp.Base(t.name)
	return base[0] < '5'
}

func (t testInfo) Name() string {
	return t.name
}

//stubs (unsued in any code so far so return zero values)
func (testInfo) ModTime() (s time.Time) { return }
func (testInfo) Sys() (s interface{})   { return }

//
// Processes
//

type testProc struct {
	t    *logger
	name string
	cmd  Command
	run  TestRun
}

func (t *testProc) Run() (error, bool) {
	t.t.Logf("%s: run dir:%s", t.name, t.cmd.Dir)
	if t.run == nil {
		return nil, true
	}
	return t.run(t.cmd)
}

//
// Test Environment World
//

type TestEnv struct {
	t     *logger
	r     *rand.Rand
	run   TestRun
	files map[string]io.ReadCloser
}

func NewTest(t *testing.T, seed int64) (w TestEnv) {
	w.t = &logger{t, []string{}}
	w.r = rand.New(rand.NewSource(seed))
	w.files = map[string]io.ReadCloser{}
	return
}

// helper functions

func (t *TestEnv) SetRun(r TestRun) {
	t.run = r
}

func (w TestEnv) Events() []string {
	return w.t.events
}

func (w TestEnv) Dump() {
	for _, ev := range w.Events() {
		w.t.t.Logf("%q", ev)
	}
}

func (w TestEnv) AddFile(name string, r io.ReadCloser) {
	w.files[name] = r
}

func (w TestEnv) Reset() {
	w.t.Reset()
}

func (w TestEnv) Logf(format string, vals ...interface{}) {
	w.t.Logf(format, vals...)
}

// implement the same functionality as the default environment

func (w TestEnv) Stat(name string) (os.FileInfo, error) {
	t := w.newInfo(name)
	w.t.Logf("world: Stat(%s): %v", name, t)
	return t, nil
}

func (w TestEnv) Readdir(name string) (fi []os.FileInfo, err error) {
	numfi := w.r.Intn(5)
	for i := 0; i < numfi; i++ {
		fi = append(fi, w.newInfo(w.randName()))
	}
	w.t.Logf("world: Readdir(%s): %v", name, fi)
	return
}

func (w TestEnv) Create(name string, mode os.FileMode) (io.WriteCloser, error) {
	w.t.Logf("world: Create(%s, %#o)", name, mode)
	return testIO{w.t, name}, nil
}

func (w TestEnv) Open(name string) (io.ReadCloser, error) {
	w.t.Logf("world: Open(%s)", name)
	if r, ok := w.files[name]; ok {
		w.t.Logf("world: returned set file")
		return r, nil
	}
	return testIO{w.t, name}, nil
}

func (w TestEnv) MkdirAll(dir string, mode os.FileMode) error {
	w.t.Logf("world: MkdirAll(%s, %#o)", dir, mode)
	return nil
}

func (w TestEnv) Make(c Command) (p Proc) {
	name := w.randName()
	w.t.Logf("world: Make(): %s: %v", name, c.Args)
	return &testProc{w.t, name, c, w.run}
}

func (w TestEnv) Exists(path string) bool {
	w.t.Logf("world: Exists(%s)", path)
	return true
}

func (w TestEnv) LookPath(path string) (string, error) {
	w.t.Logf("world: LookPath(%s)", path)
	return path, nil
}

func (w TestEnv) TempDir(prefix string) (string, error) {
	tdir := "/tmp/0" + prefix + w.randName()
	w.t.Logf("world: TempDir(%s): %s", prefix, tdir)
	return tdir, nil
}

func (w TestEnv) randName() string {
	var bytes [4]byte
	for i := range bytes {
		bytes[i] = byte(w.r.Intn(1<<8 - 1))
	}
	return fmt.Sprintf("%x", bytes)
}

func (w TestEnv) newInfo(name string) os.FileInfo {
	return testInfo{w.t, name, w.r.Int63n(1000)}
}

type TestRun func(Command) (error, bool)
