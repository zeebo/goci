package main

type State int

const (
	StateSetup = iota
	StateIdle
	StateRunning
)

func (s State) String() (r string) {
	switch s {
	case StateSetup:
		r = "Setup"
	case StateIdle:
		r = "Idle"
	case StateRunning:
		r = "Running"
	default:
		panic("unknown state")
	}
	return
}

var (
	change_state  = make(chan State)
	current_state State
)

func state_manager() {
	for current_state = range change_state {
	}
}
