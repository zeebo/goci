package main

import (
	"encoding/json"
	"github.com/zeebo/fdb"
	"time"
)

type Result struct {
	ID       int `json:"-"`
	Repo     string
	Commit   string
	Duration time.Duration
	Time     time.Time
	List     Status
	Checkout Status
	Build    Status
	Test     Status
	Clean    Status
}

type Status struct {
	Passed bool
	Output string
	Error  string
}

//Store status as a json object in the database
func (s *Status) Unserialize(p []byte) error { return json.Unmarshal(p, s) }
func (s *Status) Serialize() (p []byte)      { p, _ = json.Marshal(s); return }

//assert Status is a serializer
var _ fdb.Serializer = &Status{}
