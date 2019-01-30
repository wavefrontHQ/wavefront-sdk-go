package heartbeater

import (
	"log"
	"time"

	"github.com/wavefronthq/wavefront-sdk-go/application"
	"github.com/wavefronthq/wavefront-sdk-go/senders"
)

// Service sends a beat metric each 5 mins
type Service interface {
	Close()
}

type heartbeater struct {
	sender      senders.Sender
	application application.Tags
	source      string

	ticker *time.Ticker
	stop   chan bool
}

// Start will create and start a new Heartbeater.Service
func Start(sender senders.Sender, application application.Tags, source string) Service {
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
	error := hb.sender.SendMetric("~component.heartbeat", 1, 0, hb.source, tags)
	if error != nil {
		log.Printf("heartbeater SendMetric error: %v\n", error)
	} else {
		log.Printf("heartbeater beat OK\n")
	}
}
