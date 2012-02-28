package github

import (
	"fmt"
	"net/url"
)

func getRepo(repoPath string) (repo string) {
	parsed, err := url.Parse(repoPath)
	if err != nil {
		return
	}
	if parsed.Host != "github.com" {
		return
	}
	repo = parsed.Path[1:]

	return
}

func ClonePath(repo string) (path string, err error) {
	parsed, err := url.Parse(repo)
	if err != nil {
		return
	}
	if parsed.Host != "github.com" {
		err = fmt.Errorf("Invalid host: %q", parsed.Host)
		return
	}

	parsed.Scheme = "git"
	parsed.Path += ".git"

	path = parsed.String()
	return
}
