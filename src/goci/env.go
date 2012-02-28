package main

import "os"

type envFlag chan bool

func (i envFlag) Value() (val bool) {
	val = <-i
	i <- val
	return
}

func (i envFlag) Set(val bool) {
	select {
	case val := <-i:
		i <- val
		panic("double set")
	default:
	}
	i <- val
}

var (
	cacheDir  = os.TempDir()
	goVersion = `weekly.2012-02-22`
	goHost    = `linux-amd64`

	envInit envFlag = make(chan bool, 1)
)

func init() {
	defer envInit.Set(true)
	defer logger.Println("Base environment setup finished.")
	//set up any environment initialization here
}
