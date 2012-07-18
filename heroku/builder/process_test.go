package main

import "testing"

func TestUrlWithPath(t *testing.T) {
	a, b := urlWithPath("/a"), urlWithPath("/b")
	t.Logf("%q %q", a, b)
	if a == b {
		t.Fatal("Expected different urls")
	}
}
