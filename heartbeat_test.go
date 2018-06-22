package heartbeat

import (
	"sync"
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
	// check that we can call h(true) multiple times
	h(true)
}

func TestConcurrentAccess(t *testing.T) {
	hb := Heartbeat(time.Minute, "", nil)
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 10000; j++ {
				hb(false)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func ExampleHeartbeat_typical() {
	heartbeatExpiration := time.Second * 1
	callback := Heartbeat(heartbeatExpiration, "", nil)
	for i := 0; i < 5; i++ {
		time.Sleep(heartbeatExpiration / 2) // simulate some processing
		callback(false)                     // keep the heartbeat from expiring
	}
	callback(true) // cancel the heartbeat
}

func ExampleHeartbeat_cleanup() {
	heartbeatExpiration := time.Second * 1
	panicMessage := "Sample message"
	heartbeatFired := false
	cleanup := func() {
		if msg := recover(); msg != nil {
			// do whatever cleanup needs to be done for an expired heartbeat
			// msg will be "Sample message"
			heartbeatFired = true
		}
	}
	callback := Heartbeat(heartbeatExpiration, panicMessage, cleanup)
	time.Sleep(heartbeatExpiration * 2) // wait for the heartbeat to expire
	// heartbeatFired == true now
	callback(true) // no need to call this for an expired heartbeat, but it doesn't hurt
}

func ExampleHeartbeat_noCatch() {
	heartbeatExpiration := time.Second * 1
	panicMessage := "Sample message"
	Heartbeat(heartbeatExpiration, panicMessage, nil)
	time.Sleep(heartbeatExpiration * 2) // wait for the heartbeat to expire
	// we'll never get to here, because the heartbeat will have panic()'d
}
