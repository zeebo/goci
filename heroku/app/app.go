package main

import (
	"flag"
	"github.com/zeebo/goci/app/httputil"
	"github.com/zeebo/goci/builder"
	buweb "github.com/zeebo/goci/builder/web"
	"github.com/zeebo/goci/environ/loader"
	rudirect "github.com/zeebo/goci/runner/direct"
	ruweb "github.com/zeebo/goci/runner/web"
	"io/ioutil"
	"labix.org/v2/mgo"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
)

//env gets an environment variable with a default
func env(key, def string) (r string) {
	if r = os.Getenv(key); r == "" {
		r = def
	}
	return
}

//config stores the variables parsed by the flag package
var config struct {
	env string
}

//parse flags into our config var
func init() {
	flag.StringVar(&config.env, "env", "", "path to environment file")
	flag.Parse()
}

func main() {
	//load up the environment if its specified
	if config.env != "" {
		if err := loader.Load(config.env); err != nil {
			panic(err)
		}
	}

	//stub in a dummy database connection
	sess, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	httputil.Config.DB = sess.DB("gocitest")

	//start the server (listen first for internal requests)
	l, err := net.Listen("tcp", "0.0.0.0:"+env("PORT", "9080"))
	if err != nil {
		panic(err)
	}
	defer l.Close()
	go http.Serve(l, nil)

	var GOOS, GOARCH string
	//check if we're running direct or not
	if env("DIRECTRUN", "") == "" {
		//we're running on heroku so build for that target
		GOOS, GOARCH = "linux", "amd64"

		//create a runner
		ru := ruweb.New(
			env("APP_NAME", "goci"),
			env("API_KEY", "foo"),
			env("TRACKER", "http://goci.me/rpc/tracker"),
			env("RUNHOSTED", "http://worker.goci.me/runner/"),
		)
		http.Handle("/runner/", http.StripPrefix("/runner", ru))

		//announce the runner
		if err := ru.Announce(); err != nil {
			panic(err)
		}
		defer ru.Remove()
	} else { //we're running things directly
		GOOS, GOARCH = env("GOOS", runtime.GOOS), env("GOARCH", runtime.GOARCH)

		ru := rudirect.New(
			env("RUNPATH", "runner"),
			env("TRACKER", "http://goci.me/rpc/tracker"),
			env("RUNHOSTED", "http://worker.goci.me/runner/"),
		)
		http.Handle("/runner/", http.StripPrefix("/runner", ru))

		//announce the runner
		if err := ru.Announce(); err != nil {
			panic(err)
		}
		defer ru.Remove()
	}

	//check if we can build things
	tools := []string{"go", "hg", "bzr", "git"}
	for _, tool := range tools {
		_, err = exec.LookPath(tool)
		if tool == "git" && err != nil {
			panic("can't install git")
		}
		if err != nil {
			break
		}
	}

	//if we got any errors looking up tools, install them
	var goroot string
	if err != nil {
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
		goroot = filepath.Join(tmpdir, "go1.0.2.linux-amd64", "go")

		//add goroot/bin and venv/bin to path
		path := os.Getenv("PATH")
		path += string(filepath.ListSeparator) + filepath.Join(goroot, "bin")
		path += string(filepath.ListSeparator) + filepath.Join(tmpdir, "venv", "bin")
		os.Setenv("PATH", path)
	} else {
		//set goroot to where we found the go command
		path, err := exec.LookPath("go")
		if err != nil {
			panic("unable to find go tool")
		}

		goroot = filepath.Dir(filepath.Dir(path))
	}

	//create the builder and announce it
	bu := buweb.New(
		builder.New(GOOS, GOARCH, goroot),
		env("TRACKER", "http://goci.me/rpc/tracker"),
		env("BUILDHOSTED", "http://worker.goci.me/builder/"),
	)
	http.Handle("/builder/", http.StripPrefix("/builder", bu))

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
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)
	sig := <-ch
	log.Printf("Captured a %v\n", sig)
}
