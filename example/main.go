package main

import (
	"fmt"
	"time"
)

func main() {
	var testChan chan bool
	timer := time.NewTimer(0)
	go func() {
		time.Sleep(4 * time.Second)
		testChan = make(chan bool)
		testChan <- true
	}()
	for {
		select {
		case <-timer.C:
			fmt.Println("timer")
			timer.Reset(2 * time.Second)
		case <-testChan:
			fmt.Println(111)
		}
	}
}
