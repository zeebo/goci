package github

import "time"

type HookMessage struct {
	Before, After, Ref string
	Commits            []Commit
	Repository         Repository
}

type Repository struct {
	Name                string
	URL, Homepage       string
	Pledie, Description string
	Watchers, Forks     int
	Private             int
	Owner               Author
}

type Author struct {
	Name, Email string
}

type Commit struct {
	ID, Message              string
	Timestamp                time.Time
	URL                      string
	Added, Removed, Modified []string
	Author                   Author
}
