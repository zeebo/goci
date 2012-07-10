package rpc

import "fmt"

//Error is an error type suitable for sending over an rpc response
type Error string

func (r Error) Error() string {
	return string(r)
}

func Errorf(format string, v ...interface{}) error {
	return Error(fmt.Sprintf(format, v...))
}
