package heartbeat

import (
	"testing"
	"time"
)

func TestHeartbeat(t *testing.T) {
	d := time.Millisecond * 50
	msg := "Test message"
	fired := false
	cleanup := func() {
		if m := recover(); m != nil {
			fired = true
			if msg != m {
				t.Errorf("Override panic message not set. Expected %s, got %s\n", msg, m)
			}
		}
	}
	h := Heartbeat(d, msg, cleanup)
	time.Sleep(2 * d)
	if !fired {
		t.Errorf("Heartbeat timer didn't fire")
	}
	h = Heartbeat(d, msg, nil)
	tt := time.NewTicker(d / 2)
	i := 0
	for {
		select {
		case <-tt.C:
			i++
			if i < 4 {
				h(false)
			} else if i == 4 {
				h(true)
			} else if i == 10 {
				return
			}
		}
	}
}
