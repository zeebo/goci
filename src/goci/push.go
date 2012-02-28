package main

import (
	"encoding/json"
	"github"
	"net/http"
	"sync"
	"time"
)

var workMutex sync.Mutex

func handlePush(w http.ResponseWriter, r *http.Request) {
	var p github.HookMessage
	logger.Println("Incoming JSON:", r.FormValue("payload"))
	if err := json.Unmarshal([]byte(r.FormValue("payload")), &p); err != nil {
		errLogger.Println("json:", err)
		return
	}

	logger.Println(p)

	path, err := github.ClonePath(p.Repository.URL)
	if err != nil {
		errLogger.Println("clonePath:", err)
		return
	}

	repo := Repo(path)
	logger.Println(repo, "Cloning the repository")
	if err := repo.Clone(); err != nil {
		errLogger.Println("clone:", err)
		return
	}
	var group sync.WaitGroup
	group.Add(len(p.Commits))

	for _, commit := range p.Commits {
		go work(repo, commit.ID, group)
	}

	//launch a goroutine to clean up the repo after we're finished working
	//on it.
	go func() {
		group.Wait()
		logger.Println(repo, "Cleaning up the repository")
		repo.Cleanup()
	}()
}

func work(repo Repo, commit string, group sync.WaitGroup) {
	logger.Println(repo, commit, "Starting...")
	defer with(workMutex)()
	logger.Println(repo, commit, "Running...")
	defer logger.Println(repo, commit, "Finishing...")
	defer group.Done()

	now := time.Now()
	r := Result{
		Repo: string(repo),
	}

	defer func() {
		r.Duration = time.Now().Sub(now)
		resultsChan <- r
	}()

	logger.Println(repo, commit, "Checking out")
	if err := repo.Checkout(commit); err != nil {
		r.Checkout.Error = err.Error()
		errLogger.Println(repo, commit, "checkout:", err)
		return
	}

	r.Checkout.Passed = true

	//list
	logger.Println(repo, commit, "Listing...")
	packages, err := repo.Packages()
	if err != nil {
		r.List.Error = err.Error()
		errLogger.Println(repo, commit, "list:", err)
		return
	}

	r.List.Passed = true

	//build
	logger.Println(repo, commit, "Building...")
	stdout, stderr, err := repo.Get()
	if err != nil {
		r.Build = Status{
			Passed: false,
			Output: stdout.String(),
			Error:  stderr.String(),
		}
		errLogger.Println(repo, commit, "get:", err)
		errLogger.Printf("%+v", r.Build)
		return
	}

	r.Build = Status{
		Passed: true,
		Output: stdout.String(),
		Error:  stderr.String(),
	}

	//run a TestInstall first and ignore any errors
	logger.Println(repo, commit, "Running a test -i", packages)
	repo.TestInstall(packages)

	//test
	logger.Println(repo, commit, "Testing...")
	stdout, stderr, err = repo.Test(packages)
	if err != nil {
		r.Test = Status{
			Passed: false,
			Output: stdout.String(),
			Error:  stderr.String(),
		}
		errLogger.Println(repo, commit, "test:", err)
		errLogger.Printf("%+v", r.Test)
		return
	}

	r.Test = Status{
		Passed: true,
		Output: stdout.String(),
		Error:  stderr.String(),
	}

	logger.Println(repo, commit, "PASS")
}
