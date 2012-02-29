package main

import "testing"

func TestRingBuf(t *testing.T) {
	buf := NewResultRingBuf(2)
	if len(buf.Slice()) != 0 {
		t.Fatal("Wrong size")
	}

	buf.Push(&Result{})
	if len(buf.Slice()) != 1 {
		t.Fatal("Wrong size")
	}

	buf.Push(&Result{})
	if len(buf.Slice()) != 2 {
		t.Fatal("Wrong size")
	}

	for i := 0; i < 10; i++ {
		buf.Push(&Result{
			ID: i + 1,
		})
		sl := buf.Slice()
		if sl[0].ID != i || sl[1].ID != i+1 {
			t.Fatal("Wrong wrapping")
		}
	}
}
