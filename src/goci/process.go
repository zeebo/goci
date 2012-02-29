package main

import "time"

type workItem struct {
	repo    Repo
	commits []string
}

//a queue of 1000 items before handlers start backing up
var processQueue = make(chan workItem, 1000)

func enqueue(repo Repo, commits []string) {
	notify <- Signal{
		"repo":    repo,
		"commits": commits,
		"event":   "enqueue",
	}
	processQueue <- workItem{repo, commits}
}

func processConsumer() {
	for {
		w := <-processQueue
		process(w.repo, w.commits)
	}
}

func init() {
	go processConsumer()
}

func process(repo Repo, commits []string) {
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
	for _, commit := range commits {
		work(repo, commit)
	}
}

type Notifier struct {
	repo   Repo
	commit string
}

func newNotifier(repo Repo, commit string) *Notifier {
	return &Notifier{
		repo:   repo,
		commit: commit,
	}
}

func (n *Notifier) Notify(event string) {
	notify <- Signal{
		"repo":   n.repo,
		"commit": n.commit,
		"event":  event,
	}
}

func work(repo Repo, commit string) {
	n := newNotifier(repo, commit)
	var passed bool

	n.Notify("start")
	defer n.Notify("finish")

	defer func() {
		switch passed {
		case true:
			n.Notify("pass")
		case false:
			n.Notify("fail")
		}
	}()

	now := time.Now()
	r := Result{
		Repo:   string(repo),
		Commit: commit,
	}

	defer func() {
		r.Duration = time.Now().Sub(now)
		resultsChan <- r
	}()

	n.Notify("checkout")
	if err := repo.Checkout(commit); err != nil {
		r.Checkout.Error = err.Error()
		errLogger.Println(repo, commit, "checkout:", err)
		return
	}

	r.Checkout.Passed = true

	//list
	n.Notify("list")
	packages, err := repo.Packages()
	if err != nil {
		r.List.Error = err.Error()
		errLogger.Println(repo, commit, "list:", err)
		return
	}

	r.List.Passed = true

	//build
	n.Notify("build")
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

	//go get -v spits it's output to stderr, so log that as stdout for
	//the build phase if it passed.
	r.Build = Status{
		Passed: true,
		Output: stderr.String(),
	}

	//run a TestInstall first and ignore any errors
	n.Notify("testinstall")
	repo.TestInstall(packages)

	//test
	n.Notify("test")
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

	passed = true
}
