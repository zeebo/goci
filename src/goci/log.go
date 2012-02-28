package main

import (
	"log"
	"os"
)

type Logger interface {
	Fatal(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

//set up our apps loggers
var (
	logger    Logger = log.New(os.Stdout, "app: ", log.Lshortfile)
	errLogger Logger = log.New(os.Stderr, "error: ", log.Lshortfile)
)
