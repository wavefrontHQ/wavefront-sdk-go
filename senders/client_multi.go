package senders

import (
	"github.com/wavefronthq/wavefront-sdk-go/event"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
)

// MultiSender Interface for sending metrics, distributions and spans to multiple Wavefronts at the same time
type MultiSender interface {
	Sender
	Add(s Sender)
}

type multiSender struct {
	senders []Sender
}

// NewMultiSender creates a new Wavefront MultiClient
func NewMultiSender() MultiSender {
	ms := &multiSender{}
	return ms
}

func (ms *multiSender) Add(s Sender) {
	ms.senders = append(ms.senders, s)
}

func (ms *multiSender) SendMetric(name string, value float64, ts int64, source string, tags map[string]string) error {
	var firstErr error
	for _, sender := range ms.senders {
		err := sender.SendMetric(name, value, ts, source, tags)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (ms *multiSender) SendDeltaCounter(name string, value float64, source string, tags map[string]string) error {
	var firstErr error
	for _, sender := range ms.senders {
		err := sender.SendDeltaCounter(name, value, source, tags)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (ms *multiSender) SendDistribution(name string, centroids []histogram.Centroid, hgs map[histogram.Granularity]bool, ts int64, source string, tags map[string]string) error {
	var firstErr error
	for _, sender := range ms.senders {
		err := sender.SendDistribution(name, centroids, hgs, ts, source, tags)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (ms *multiSender) SendSpan(name string, startMillis, durationMillis int64, source, traceID, spanID string, parents, followsFrom []string, tags []SpanTag, spanLogs []SpanLog) error {
	var firstErr error
	for _, sender := range ms.senders {
		err := sender.SendSpan(name, startMillis, durationMillis, source, traceID, spanID, parents, followsFrom, tags, spanLogs)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (ms *multiSender) SendEvent(name string, startMillis, endMillis int64, source string, tags map[string]string, setters ...event.Option) error {
	var firstErr error
	for _, sender := range ms.senders {
		err := sender.SendEvent(name, startMillis, endMillis, source, tags, setters...)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (ms *multiSender) Flush() error {
	var firstErr error
	for _, sender := range ms.senders {
		err := sender.Flush()
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (ms *multiSender) GetFailureCount() int64 {
	var fc int64
	for _, sender := range ms.senders {
		fc += sender.GetFailureCount()
	}
	return fc
}

func (ms *multiSender) Start() {
	for _, sender := range ms.senders {
		sender.Start()
	}
}

func (ms *multiSender) Close() {
	for _, sender := range ms.senders {
		sender.Close()
	}
}

//TODO: remove, just for tesing
func (ms *multiSender) SetSource(string) {}
