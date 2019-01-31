package application

import (
	"log"
	"time"

	"github.com/wavefronthq/wavefront-sdk-go/senders"
)

// HeartbeatService sends a heartbeat metric each 5 mins
type HeartbeatService interface {
	Close()
}

type heartbeater struct {
	sender      senders.Sender
	application Tags
	source      string

	ticker *time.Ticker
	stop   chan bool
}

// StartHeartbeatService will create and start a new HeartbeatService
func StartHeartbeatService(sender senders.Sender, application Tags, source string) HeartbeatService {
	hb := &heartbeater{
		sender:      sender,
		application: application,
		source:      source,
		ticker:      time.NewTicker(5 * time.Minute),
		stop:        make(chan bool, 1),
	}

	go func() {
		for {
			select {
			case <-hb.ticker.C:
				hb.beat()
			case <-hb.stop:
				return
			}
		}
	}()

	hb.beat()

	return hb
}

func (hb *heartbeater) Close() {
	hb.stop <- true
}

func (hb *heartbeater) beat() {
	tags := hb.application.Map()
	tags["component"] = "wavefront-generated"
	err := hb.sender.SendMetric("~component.heartbeat", 1, 0, hb.source, tags)
	if err != nil {
		log.Printf("heartbeater SendMetric error: %v\n", err)
	}
}
