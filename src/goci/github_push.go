package main

import (
	"encoding/json"
	"github"
	"net/http"
	"strings"
)

var githubProcessChannel = make(chan github.HookMessage)

func init() {
	//run our consumer
	go githubProcessConsumer()
}

func githubProcessConsumer() {
	for p := range githubProcessChannel {
		commits := make([]string, len(p.Commits))
		for i, commit := range p.Commits {
			commits[i] = commit.ID
		}
		logger.Printf("[github] %s. Commits: %v", p.Repository.URL, commits)

		path, err := github.ClonePath(p.Repository.URL)
		if err != nil {
			errLogger.Println("clonePath:", err)
			return
		}

		enqueue(Repo(path), commits)
	}
}

func handleGithubPush(w http.ResponseWriter, r *http.Request) {
	//avoid an alloc of the whole message
	dec := json.NewDecoder(strings.NewReader(r.FormValue("payload")))

	//decode
	var p github.HookMessage
	if err := dec.Decode(&p); err != nil {
		errLogger.Println("json:", err)
		return
	}

	githubProcessChannel <- p
}
