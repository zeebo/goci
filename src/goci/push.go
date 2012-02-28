package main

import (
	"encoding/json"
	"github"
	"net/http"
	"sync"
)

var workMutex sync.Mutex

func handlePush(w http.ResponseWriter, r *http.Request) {
	var p github.HookMessage
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
		repo.Cleanup()
	}()
}

func work(repo Repo, commit string, group sync.WaitGroup) {
	defer with(workMutex)()
	defer group.Done()

	if err := repo.Checkout(commit); err != nil {
		errLogger.Println(repo, commit, "checkout:", err)
		return
	}

	//build
	stdout, stderr, err := repo.Get()
	if err != nil {
		errLogger.Println(repo, commit, "get:", err)
		errLogger.Println("stdout:", stdout.String())
		errLogger.Println("stderr:", stderr.String())
		return
	}

	//test
	stdout, stderr, err = repo.Test()
	if err != nil {
		errLogger.Println(repo, commit, "test:", err)
		errLogger.Println("stdout:", stdout.String())
		errLogger.Println("stderr:", stderr.String())
		return
	}

	logger.Println(repo, commit, "PASS")
	logger.Println(repo, commit, stdout.String())
}
