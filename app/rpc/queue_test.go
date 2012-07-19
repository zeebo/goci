package rpc

import (
	"reflect"
	"testing"
)

func TestQueue(t *testing.T) {
	val := struct{ X int }{6}
	num := 5
	q := newQueue()

	//push num in
	for i := 0; i < num; i++ {
		q.push(val)
	}

	//read num out
	for i := 0; i < num; i++ {
		if got := q.pop(); !reflect.DeepEqual(val, got) {
			t.Fatal("Expected %#v. Got %#v", val, got)
		}
	}
}
