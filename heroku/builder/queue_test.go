package main

import (
	"github.com/zeebo/goci/app/rpc"
	"reflect"
	"testing"
)

func TestTaskQueue(t *testing.T) {
	val := rpc.BuilderTask{Runner: "foo"}
	num := 5
	q := newTaskQueue()

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
