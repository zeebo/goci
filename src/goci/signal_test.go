package main

import "testing"

func TestBroadcast(t *testing.T) {
	one, two := make(SignalPipe), make(SignalPipe)
	signalRegister <- one
	signalRegister <- two

	notify <- nil

	if g := <-one; g != nil {
		t.Fatal("Expected %v Got %v", nil, g)
	}
	if g := <-two; g != nil {
		t.Fatal("Expected %v Got %v", nil, g)
	}

	signalUnregister <- one
	notify <- nil

	if g := <-two; g != nil {
		t.Fatal("Expected %v Got %v", nil, g)
	}
	select {
	case <-one:
		t.Fatal("Got value after unregister")
	default:
	}
}
