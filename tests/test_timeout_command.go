package main

import (
	"fmt"
	"os/exec"
	"time"
)

func timeout(cmd *exec.Cmd, dur time.Duration) (ok bool) {
	done := make(chan bool, 1)
	if err := cmd.Start(); err != nil {
		return
	}
	defer cmd.Process.Kill()

	//start a race
	go func() {
		cmd.Wait()
		done <- true
	}()

	go func() {
		<-time.After(dur)
		done <- false
	}()

	//see who won
	ok = <-done
	return
}

func main() {
	cmd := exec.Command("sleep", "500")
	now, ret := time.Now(), timeout(cmd, 3*time.Second)
	fmt.Println(ret, time.Since(now))

	cmd = exec.Command("sleep", "5")
	now, ret = time.Now(), timeout(cmd, 7*time.Second)
	fmt.Println(ret, time.Since(now))
}
