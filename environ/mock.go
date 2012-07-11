package environ

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"testing"
	"time"
	fp "path/filepath"
)

//
// Logger
//

type logger struct {
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

//TestEnv is an environment that records all of the actions taken on it for 
//later inspection.
type TestEnv struct {
	t     *logger
	r     *rand.Rand
	run   TestRun
	files map[string]io.ReadCloser
}

//NewTest sets up a new testing environment with the given seed.
func NewTest(seed int64) (w TestEnv) {
	w.t = &logger{[]string{}}
	w.r = rand.New(rand.NewSource(seed))
	w.files = map[string]io.ReadCloser{}
	return
}

// helper functions

//SetRun sets the function that Procs created by the Make method run.
func (t *TestEnv) SetRun(r TestRun) {
	t.run = r
}

//AddFile adds the named file to the environment so that if it is opened it will
//return the given io.ReadCloser
func (w TestEnv) AddFile(name string, r io.ReadCloser) {
	w.files[name] = r
}

//Events returns the list of events taken on this environment.
func (w TestEnv) Events() []string {
	return w.t.events
}

//Dump prints all the events this environment has seen to the testing logger.
func (w TestEnv) Dump(t *testing.T) {
	for _, ev := range w.Events() {
		t.Logf("%q", ev)
	}
}

//Reset drops all the events.
func (w TestEnv) Reset() {
	w.t.Reset()
}

//Logf logs the given event to the stream of events.
func (w TestEnv) Logf(format string, vals ...interface{}) {
	w.t.Logf(format, vals...)
}

// implement the same functionality as the default environment

//Stat logs the call and returns an os.FileInfo that also logs calls.
func (w TestEnv) Stat(name string) (os.FileInfo, error) {
	t := w.newInfo(name)
	w.t.Logf("world: Stat(%s): %v", name, t)
	return t, nil
}

//Readdir logs the call and returns a random set of up to 5 os.FileInfos that
//also log calls. It does not return the same thing for multiple calls.
func (w TestEnv) Readdir(name string) (fi []os.FileInfo, err error) {
	numfi := w.r.Intn(5)
	for i := 0; i < numfi; i++ {
		fi = append(fi, w.newInfo(w.randName()))
	}
	w.t.Logf("world: Readdir(%s): %v", name, fi)
	return
}

//Create logs the call and returns an io.WriteCloser that logs the number of
//bytes written to it and when it is closed.
func (w TestEnv) Create(name string, mode os.FileMode) (io.WriteCloser, error) {
	w.t.Logf("world: Create(%s, %#o)", name, mode)
	return testIO{w.t, name}, nil
}

//Open logs the call and returns an io.ReadCloser that logs the number of bytes
//read from it and when it is closed.
func (w TestEnv) Open(name string) (io.ReadCloser, error) {
	w.t.Logf("world: Open(%s)", name)
	if r, ok := w.files[name]; ok {
		w.t.Logf("world: returned set file")
		return r, nil
	}
	return testIO{w.t, name}, nil
}

//MkdirAll logs the call.
func (w TestEnv) MkdirAll(dir string, mode os.FileMode) error {
	w.t.Logf("world: MkdirAll(%s, %#o)", dir, mode)
	return nil
}

//Make logs the call and returns a Proc that runs the function set by SetRun
//when called, logging that call.
func (w TestEnv) Make(c Command) (p Proc) {
	name := w.randName()
	w.t.Logf("world: Make(): %s: %v", name, c.Args)
	return &testProc{w.t, name, c, w.run}
}

//Exists logs the call and returns true.
func (w TestEnv) Exists(path string) bool {
	w.t.Logf("world: Exists(%s)", path)
	return true
}

//LookPath logs the call and returns what was passed in with no error.
func (w TestEnv) LookPath(path string) (string, error) {
	w.t.Logf("world: LookPath(%s)", path)
	return path, nil
}

//TempDir logs the call and returns a random directory with the given prefix.
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

//TestRun is a function that will be called when a Proc executes when added
//to a TestEnv.
type TestRun func(Command) (error, bool)
