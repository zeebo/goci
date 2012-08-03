package main

import (
	"github.com/zeebo/goci/builder"
	buweb "github.com/zeebo/goci/builder/web"
	ruweb "github.com/zeebo/goci/runner/web"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
)

//env gets an environment variable with a default
func env(key, def string) (r string) {
	if r = os.Getenv(key); r == "" {
		r = def
	}
	return
}

func main() {
	ru := ruweb.New(
		env("APP_NAME", "goci"),
		env("API_KEY", "foo"),
		env("TRACKER", "http://goci.me/tracker"),
		env("RUNHOSTED", "http://runner.goci.me/runner"),
	)
	http.Handle("/runner", http.StripPrefix("/runner", ru))

	//start the server
	go http.ListenAndServe(":"+env("PORT", "9080"), nil)

	//announce the runner
	if err := ru.Announce(); err != nil {
		panic(err)
	}
	defer ru.Remove()

	//run the setup script
	tmpdir, err := ioutil.TempDir("", "tools")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpdir)

	cmd := exec.Command("bash", "heroku_setup.sh", "heroku/dist", tmpdir)
	if err := cmd.Run(); err != nil {
		panic(err)
	}

	//store our goroot
	goroot := filepath.Join(tmpdir, "go1.0.2.linux-amd64", "go")

	//add goroot/bin and venv/bin to path
	path := os.Getenv("PATH")
	path += string(filepath.ListSeparator) + filepath.Join(goroot, "bin")
	path += string(filepath.ListSeparator) + filepath.Join(tmpdir, "venv", "bin")
	os.Setenv("PATH", path)

	//create the builder and announce it
	bu := buweb.New(
		builder.New("linux", "amd64", goroot),
		env("TRACKER", "http://goci.me/tracker"),
		env("BUILDHOSTED", "http://runner.goci.me/builder"),
	)
	http.Handle("/builder", http.StripPrefix("/builder", bu))

	if err := bu.Announce(); err != nil {
		panic(err)
	}
	defer bu.Remove()

	//wait for a signal
	signals := []os.Signal{
		syscall.SIGQUIT,
		syscall.SIGKILL,
		syscall.SIGINT,
		syscall.SIGTERM,
	}
	ch := make(chan os.Signal, len(signals))
	signal.Notify(ch, signals...)
	sig := <-ch
	log.Printf("Captured a %v\n", sig)
}
