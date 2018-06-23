// Package heartbeat provides functionality to invoke panic() if the program does not call a heartbeat function within a configurable time period.
package heartbeat

import (
	"time"
)

type heartbeatStatus int

const (
	expired heartbeatStatus = iota
	cancelled
)

// Heartbeat launches a go routine that will panic if the returned function is not called at least once per time interval t.
// message is an optional parameter. If supplied, it will override the default panic message "Heartbeat timer expired."
// cleanup is an optional parameter. If supplied, it will be called during the panic() unwinding. It can be used to recover() from the panic, if the program does not want to exit.
func Heartbeat(t time.Duration, message string, cleanup func()) func(cancel bool) {
	if len(message) == 0 {
		message = "Heartbeat timer expired."
	}
	return heartbeat(t, true, func(status heartbeatStatus) {
		if cleanup != nil {
			defer cleanup()
		}
		if status == expired {
			panic(message)
		}
	})
}

// HeartbeatMonitor launches a go routine that will call a timeout handler function if the returned function is not called at least once per time interval t.
// This process does not cause a panic and will repeat until the function is canceled
func HeartbeatMonitor(duration time.Duration, heartbeatTimeoutHandler func()) func(cancel bool) {
	return heartbeat(duration, false, func(status heartbeatStatus) {
		if status == expired && heartbeatTimeoutHandler != nil {
			heartbeatTimeoutHandler()
		}
	})
}

// Channel sends true on the returned channel if the returned function is not called at
// least once per the supplied time interval. To cancel the heartbeat, call the returned
// function with true.
func Channel(t time.Duration) (<-chan bool, func(bool)) {
	expired := make(chan bool)
	return expired, heartbeat(t, true, func(status heartbeatStatus) {
		expired <- true
		close(expired)
	})
}

func heartbeat(t time.Duration, onetime bool, cb func(heartbeatStatus)) func(cancel bool) {
	hb := make(chan bool)
	go func() {
		for {
			select {
			case cancel := <-hb:
				if cancel {
					cb(cancelled)
					return
				}
			case <-time.After(t):
				cb(expired)
				if onetime {
					return
				}
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
