package main

import "time"

type workItem struct {
	Repo    Repo
	Commits []string
}

//a queue of 100 items before handlers start backing up
const maxQueueSize = 100

var (
	processQueue = make(chan workItem)
	currentQueue = make(chan []workItem)
)

func enqueue(repo Repo, commits []string) {
	processQueue <- workItem{repo, commits}
	notify <- Signal{
		"repo":    repo,
		"commits": commits,
		"event":   "enqueue",
	}
}

func processConsumer(in chan workItem) {
	for {
		w := <-in
		process(w.Repo, w.Commits)
	}
}

func processQueuer(in chan workItem, current chan []workItem) (out chan workItem) {
	out = make(chan workItem)
	go func() {
		processQueueBuf := make(chan workItem, maxQueueSize)
		buf := make([]workItem, 0)
		var val *workItem
		for {
			if val == nil {
				select {
				case w := <-in:
					processQueueBuf <- w
					buf = append(buf, w)
				case temp := <-processQueueBuf:
					val = &temp
				case current <- buf:
				}
			} else {
				select {
				case w := <-in:
					processQueueBuf <- w
					buf = append(buf, w)
				case out <- *val:
					val = nil
					buf = buf[1:]
				case current <- buf:
				}
			}
		}
	}()
	return out
}

func init() {
	mid := processQueuer(processQueue, currentQueue)
	go processConsumer(mid)
}

func process(repo Repo, commits []string) {
	defer func() {
		notify <- Signal{
			"repo":  repo,
			"event": "processfinished",
		}
	}()

	notify <- Signal{
		"repo":  repo,
		"event": "repoclean",
	}
	repo.Cleanup()

	notify <- Signal{
		"repo":  repo,
		"event": "repoclone",
	}
	if err := repo.Clone(); err != nil {
		errLogger.Println("clone:", err)
		return
	}
	defer func() {
		notify <- Signal{
			"repo":  repo,
			"event": "repoclean",
		}
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
		Time:   now,
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
	stdout, stderr, err := repo.Get(packages)
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

	//defer a clean
	defer func() {
		n.Notify("clean")
		stdout, stderr, err := repo.Clean(packages)
		if err != nil {
			r.Clean = Status{
				Passed: false,
				Output: stdout.String(),
				Error:  stderr.String(),
			}
			errLogger.Println(repo, commit, "clean:", err)
			errLogger.Printf("%+v", r.Clean)
			return
		}

		r.Clean = Status{
			Passed: true,
			Output: stdout.String(),
		}
	}()

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
