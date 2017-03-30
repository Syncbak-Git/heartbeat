// Package heartbeat provides functionality to invoke panic() if the program does not call a heartbeat function within a configurable time period.
package heartbeat

import (
	"time"
)

// Heartbeat launches a go routine that will panic if the returned function is not called at least once per time interval t.
// message is an optional parameter. If supplied, it will override the default panic message "Heartbeat timer expired."
// cleanup is an optional parameter. If supplied, it will be called during the panic() unwinding. It can be used to recover() from the panic, if the program does not want to exit.
func Heartbeat(t time.Duration, message string, cleanup func()) func(cancel bool) {
	if len(message) == 0 {
		message = "Heartbeat timer expired."
	}
	tt := time.NewTimer(t)
	quit := make(chan interface{})
	f := func(cancel bool) {
		if cancel {
			tt.Stop()
			close(quit)
		} else {
			tt.Reset(t)
		}
	}
	go func() {
		if cleanup != nil {
			defer cleanup()
		}
		for {
			select {
			case <-quit:
				return
			case <-tt.C:
				panic(message)
			}
		}
	}()
	return f
}

// Channel sends true on the returned channel if the returned function is not called at
// least once per the supplied time interval. To cancel the heartbeat, call the returned
// function with true.
func Channel(t time.Duration) (<-chan bool, func(bool)) {
	tt := time.NewTimer(t)
	quit := make(chan interface{})
	expired := make(chan bool, 1)
	f := func(cancel bool) {
		if cancel {
			tt.Stop()
			close(quit)
		} else {
			tt.Reset(t)
		}
	}
	go func() {
		select {
		case <-quit:
		case <-tt.C:
		}
		expired <- true
	}()
	return expired, f
}
