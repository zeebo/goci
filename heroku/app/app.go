package main

import (
	"flag"
	"github.com/zeebo/goci/app/httputil"
	"github.com/zeebo/goci/app/rpc/router"
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

//Service can handle http requests and announce and remove its presence to a
//tracker.
type Service interface {
	http.Handler
	Announce() error
	Remove() error
}

//newWebRunner returns a service for running tests on the heroku dyno mesh.
func newWebRunner() Service {
	//create a runner
	ru := ruweb.New(
		env("APP_NAME", "goci"),
		env("API_KEY", "foo"),
		httputil.Absolute(router.Lookup("Tracker")),
		httputil.Absolute("/runner/"),
	)
	return ru
}

//newDirectRunner returns a service for running tests on the local machine.
func newDirectRunner() Service {
	ru := rudirect.New(
		env("RUNPATH", "runner"),
		httputil.Absolute(router.Lookup("Tracker")),
		httputil.Absolute("/runner/"),
	)
	return ru
}

//checkTools checks for the presence of all the tools we need to build
func checkTools() (err error) {
	tools := []string{"go", "hg", "bzr", "git"}
	for _, tool := range tools {
		_, err = exec.LookPath(tool)
		if tool == "git" && err != nil {
			panic("can't install git")
		}
		if err != nil {
			return
		}
	}
	return
}

func main() {
	//load up the environment if its specified
	if config.env != "" {
		if err := loader.Load(config.env); err != nil {
			panic(err)
		}
	}

	//connect to the mongo database.
	sess, err := mgo.Dial(env("DATABASE", "mongodb://localhost/gocitest"))
	if err != nil {
		panic(err)
	}
	//empty implies whatever was specified in dial.
	httputil.Config.DB = sess.DB("")

	//set up the httputil domain so we can build absolute urls
	httputil.Config.Domain = env("DOMAIN", "localhost:9080")

	//start the server (listen first for internal requests)
	l, err := net.Listen("tcp", "0.0.0.0:"+env("PORT", "9080"))
	if err != nil {
		panic(err)
	}
	defer l.Close()
	go http.Serve(l, nil)

	//set up some vars for our target os and arch and the runner
	var GOOS, GOARCH string
	var runner Service

	//check if we're running direct or not
	if env("DIRECTRUN", "") == "" {
		//we're running on heroku so build for that target
		GOOS, GOARCH = "linux", "amd64"
		runner = newWebRunner()
	} else {
		//we're running things directly
		GOOS, GOARCH = env("GOOS", runtime.GOOS), env("GOARCH", runtime.GOARCH)
		runner = newDirectRunner()
	}

	//add the runner to our system
	http.Handle("/runner/", http.StripPrefix("/runner", runner))

	//announce the runner
	if err := runner.Announce(); err != nil {
		panic(err)
	}
	defer runner.Remove()

	var goroot string
	if err := checkTools(); err != nil {
		//create a temporary directory to house the go tool and hg+bzr.
		tmpdir, err := ioutil.TempDir("", "tools")
		if err != nil {
			panic(err)
		}
		defer os.RemoveAll(tmpdir)

		//execute our setup script
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

		//check for the tools again as they should installed
		if err := checkTools(); err != nil {
			panic("couldn't find tools after installing: " + err.Error())
		}
	}

	//if we don't have goroot set yet set it to where we find the go command
	if goroot == "" {
		path, err := exec.LookPath("go")
		if err != nil {
			panic("unable to find go tool")
		}
		goroot = filepath.Dir(filepath.Dir(path))
	}

	//create the builder and announce it
	bu := buweb.New(
		builder.New(GOOS, GOARCH, goroot),
		httputil.Absolute(router.Lookup("Tracker")),
		httputil.Absolute("/builder/"),
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
