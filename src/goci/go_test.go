package main

import "testing"

func TestRoot(t *testing.T) {
	path, err := Root("weekly.2012-02-22")
	if err != nil {
		t.Error("root:", err)
	}
	t.Log(path)
}
