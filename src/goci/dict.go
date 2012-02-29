package main

import "encoding/json"

//d and l are simple types for creating json lists and dictionaries
type d map[string]interface{}
type l []interface{}

func (d d) String() string {
	val, _ := json.Marshal(d)
	return string(val)
}

func (l l) String() string {
	val, _ := json.Marshal(l)
	return string(val)
}
