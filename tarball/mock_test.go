package tarball

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"testing"
	"time"
	fp "path/filepath"
)

type logger struct {
	t      *testing.T
	events []string
}

func (l *logger) Logf(format string, vals ...interface{}) {
	s := fmt.Sprintf(format, vals...)
	// l.t.Logf("%q", s)
	l.events = append(l.events, s)
}

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
func (t testInfo) Name() string { return t.name }

//stubs (unsued in our code)
func (testInfo) ModTime() (s time.Time) { panic("unused") }
func (testInfo) Sys() (s interface{})   { panic("unused") }

type testWorld struct {
	t *logger
	r *rand.Rand

	files map[string]io.ReadCloser
}

func (w testWorld) events() []string {
	return w.t.events
}

func newTestWorld(t *testing.T, seed int64) (w testWorld) {
	w.t = &logger{t, []string{}}
	w.r = rand.New(rand.NewSource(seed))
	w.files = map[string]io.ReadCloser{}
	return
}

func (w testWorld) Stat(name string) (os.FileInfo, error) {
	t := w.newInfo(name)
	w.t.Logf("world: Stat(%s): %v", name, t)
	return t, nil
}

func (w testWorld) Readdir(name string) (fi []os.FileInfo, err error) {
	numfi := w.r.Intn(5)
	for i := 0; i < numfi; i++ {
		fi = append(fi, w.newInfo(w.randName()))
	}
	w.t.Logf("world: Readdir(%s): %v", name, fi)
	return
}

func (w testWorld) Create(name string, mode os.FileMode) (io.WriteCloser, error) {
	w.t.Logf("world: Create(%s, %#o)", name, mode)
	return testIO{w.t, name}, nil
}

func (w testWorld) Open(name string) (io.ReadCloser, error) {
	w.t.Logf("world: Open(%s)", name)
	if r, ok := w.files[name]; ok {
		w.t.Logf("world: returned set file")
		return r, nil
	}
	return testIO{w.t, name}, nil
}

func (w testWorld) MkdirAll(dir string, mode os.FileMode) error {
	w.t.Logf("world: MkdirAll(%s, %#o)", dir, mode)
	return nil
}

func (w testWorld) randName() string {
	var bytes [4]byte
	for i := range bytes {
		bytes[i] = byte(w.r.Intn(1<<8 - 1))
	}
	return fmt.Sprintf("%x", bytes)
}

func (w testWorld) newInfo(name string) os.FileInfo {
	return testInfo{w.t, name, w.r.Int63n(1000)}
}
