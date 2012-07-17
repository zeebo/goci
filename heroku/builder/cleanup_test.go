package main

import "testing"

func TestCleaner(t *testing.T) {
	c := newCleaner()
	var ok bool
	c.attach(func() {
		ok = true
	})
	c.cleanup()
	if !ok {
		t.Fatal("Function not run")
	}
}
