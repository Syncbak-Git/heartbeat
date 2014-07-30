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
	f := func(cancel bool) {
		if cancel {
			tt.Stop()
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
			case <-tt.C:
				panic(message)
			}
		}
	}()
	return f
}
