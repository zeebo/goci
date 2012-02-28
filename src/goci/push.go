package main

import (
	"encoding/json"
	"github"
	"net/http"
	"time"
)

var processChannel = make(chan github.HookMessage)

func init() {
	//wait for our environment init
	_ = envInit.Value()

	//run our consumer
	go processConsumer()
}

func processConsumer() {
	for p := range processChannel {
		process(p)
	}
}

func process(p github.HookMessage) {
	commits := make([]string, len(p.Commits))
	for i, commit := range p.Commits {
		commits[i] = commit.ID
	}
	logger.Printf("Pull for %s. Commits: %v", p.Repository.URL, commits)

	path, err := github.ClonePath(p.Repository.URL)
	if err != nil {
		errLogger.Println("clonePath:", err)
		return
	}

	repo := Repo(path)
	logger.Println(repo, "Cleaning out any old repository")
	repo.Cleanup()

	logger.Println(repo, "Cloning the repository")
	if err := repo.Clone(); err != nil {
		errLogger.Println("clone:", err)
		return
	}
	defer func() {
		logger.Println(repo, "Cleaning up the repository")
		repo.Cleanup()
	}()

	//spawn workers
	for _, commit := range p.Commits {
		work(repo, commit.ID)
	}
}

func handlePush(w http.ResponseWriter, r *http.Request) {
	var p github.HookMessage
	if err := json.Unmarshal([]byte(r.FormValue("payload")), &p); err != nil {
		errLogger.Println("json:", err)
		return
	}
	//async send it in
	go func() {
		processChannel <- p
	}()
}

func work(repo Repo, commit string) {
	logger.Println(repo, commit, "Starting...")
	defer logger.Println(repo, commit, "Finishing...")

	now := time.Now()
	r := Result{
		Repo:   string(repo),
		Commit: commit,
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
