package appinstance

import (
	"sync"
	"time"
)

// RaiseRunningInstance brings the live Notepad window forward when a forwarded
// open arrives (set from main after InitWindow).
var RaiseRunningInstance func()

func startPendingPoll(onOpen func(path string)) (stop func()) {
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(120 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				path := consumePendingOpen()
				if path == "" {
					continue
				}
				if RaiseRunningInstance != nil {
					RaiseRunningInstance()
				}
				if onOpen != nil {
					onOpen(path)
				}
			}
		}
	}()
	return func() {
		close(done)
		wg.Wait()
	}
}
