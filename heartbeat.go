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
	hb := make(chan bool)
	go func() {
		if cleanup != nil {
			defer cleanup()
		}
		for {
			select {
			case cancel := <-hb:
				if cancel {
					return
				}
			case <-time.After(t):
				panic(message)
			}
		}
	}()
	return func(cancel bool) {
		select {
		case hb <- cancel:
		default:
		}
	}
}

// HeartbeatMonitor launches a go routine that will call a timeout handler function if the returned function is not called at least once per time interval t.
// This process does not cause a panic and will repeat until the function is canceled
func HeartbeatMonitor(duration time.Duration, heartbeatTimeoutHandler func()) func(cancel bool) {

	tt := time.NewTimer(duration)

	quit := make(chan interface{})

	// This is the callback function that is returned
	f := func(cancel bool) {
		if cancel {
			tt.Stop()
			close(quit)
		} else {
			tt.Reset(duration)
		}
	}

	go func() {

		for {
			select {
			case <-quit:
				return
			case <-tt.C:
				if heartbeatTimeoutHandler != nil {
					heartbeatTimeoutHandler()
				}
				tt.Reset(duration)
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
		close(expired)
	}()
	return expired, f
}
