package lines

import (
	"errors"
	"time"
)

var throttledSleepDuration = time.Second * 30
var errThrottled = errors.New("error: throttled event creation")

type FlusherWithBackgroundTicker struct {
	buffer   chan string
	done     chan struct{}
	ticker   *time.Ticker
	interval time.Duration
	handler  Handler
	lockIfServerAsksToThrottle bool
}

func (f *FlusherWithBackgroundTicker) Stop() {
	f.ticker.Stop()
	f.done <- struct{}{}
}

func NewFlusherWithBackgroundTicker(buffer chan string, interval time.Duration, handler Handler) Flusher {
	return &FlusherWithBackgroundTicker{
		buffer:   buffer,
		interval: interval,
		handler:  handler,
	}
}

func (f *FlusherWithBackgroundTicker) Start() {
	f.done = make(chan struct{})
	f.ticker = time.NewTicker(f.interval)

	go func() {
		for {
			select {
			case <-f.ticker.C:
				err := f.handler.Flush()
				if err != nil {
					// TODO: this wants the handler's mutex, but I'm not sure why
					// isn't it just trying to back off because it's being throttled?
					// should we not just reschedule the timer until the next sleep time, and then reset the interval and resume?
					if err == errThrottled && f.lockIfServerAsksToThrottle {
						go func() {
							//lh.mtx.Lock()
							//atomic.AddInt64(&lh.throttled, 1)
							//log.Printf("sleeping for %v, buffer size: %d\n", throttledSleepDuration, len(lh.buffer))
							//time.Sleep(throttledSleepDuration)
							//lh.mtx.Unlock()
						}()
					}
				}
			case <-f.done:
				return
			}
		}
	}()
}
