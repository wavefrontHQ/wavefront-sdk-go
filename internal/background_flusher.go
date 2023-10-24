package internal

import (
	"log"
	"time"
)

type BackgroundFlusher interface {
	Start()
	Stop()
}

type backgroundFlusher struct {
	ticker   *time.Ticker
	interval time.Duration
	handler  LineHandler
	stop     chan struct{}
}

func NewBackgroundFlusher(interval time.Duration, handler LineHandler) BackgroundFlusher {
	return &backgroundFlusher{
		interval: interval,
		handler:  handler,
		stop:     make(chan struct{}),
	}
}

func (f *backgroundFlusher) Start() {
	format := f.handler.Format()
	if f.ticker != nil {
		return
	}
	f.ticker = time.NewTicker(f.interval)
	go func() {
		for {
			select {
			case tick := <-f.ticker.C:
				log.Printf("%s -- flushing at: %s\n", format, tick)
				err := f.handler.FlushWithThrottling()
				if err != nil {
					log.Printf("%s -- error during background flush: %s\n", format, err.Error())
				} else {
					log.Printf("%s -- flush completed at %s\n", format, time.Now())
				}
			case <-f.stop:
				return
			}
		}
	}()
}

func (f *backgroundFlusher) Stop() {
	f.ticker.Stop()
	f.stop <- struct{}{}
}
