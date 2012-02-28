package main

import "os"

type envFlag chan bool

func (i envFlag) IsDone() bool {
	select {
	case <-i:
		i <- true
		return true
	default:
	}
	return false
}

func (i envFlag) Wait() {
	<-i
	i <- true
}

func (i envFlag) Finished() {
	if !i.IsDone() {
		i <- true
	}
}

var (
	cacheDir, _ = os.Getwd()
	goVersion   = `weekly.2012-02-22`
	goHost      = `linux-amd64`

	envInit envFlag = make(chan bool, 1)
)

func init() {
	defer envInit.Finished()

	//set up any environment initialization here
}
