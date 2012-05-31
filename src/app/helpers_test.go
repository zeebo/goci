package main

import "testing"

func TestNewID(t *testing.T) {
	t.Log(new_id())
}

func BenchmarkNewID(b *testing.B) {
	var y string
	for i := 0; i < b.N; i++ {
		y = new_id()
	}
	_ = y
}
