package wavefront

import (
	"github.com/wavefronthq/wavefront-sdk-go/event"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
	"github.com/wavefronthq/wavefront-sdk-go/types"
)

type MultiClient interface {
	Client
	Add(wf Client)
}

type multiClient struct {
	clients []Client
}

func NewMultiClient(wfs ...Client) MultiClient {
	mc := &multiClient{}
	for _, wf := range wfs {
		mc.Add(wf)
	}
	return mc
}

func (mc *multiClient) Add(wf Client) {
	mc.clients = append(mc.clients, wf)
}

func (mc *multiClient) SendMetric(name string, value float64, ts int64, source string, tags map[string]string) error {
	var firstErr error
	for _, wf := range mc.clients {
		err := wf.SendMetric(name, value, ts, source, tags)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (mc *multiClient) SendDeltaCounter(name string, value float64, source string, tags map[string]string) error {
	var firstErr error
	for _, wf := range mc.clients {
		err := wf.SendDeltaCounter(name, value, source, tags)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (mc *multiClient) SendDistribution(name string, centroids []histogram.Centroid, hgs map[histogram.Granularity]bool, ts int64, source string, tags map[string]string) error {
	var firstErr error
	for _, wf := range mc.clients {
		err := wf.SendDistribution(name, centroids, hgs, ts, source, tags)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (mc *multiClient) SendSpan(name string, startMillis, durationMillis int64, source, traceID, spanID string, parents, followsFrom []string, tags []types.SpanTag, spanLogs []types.SpanLog) error {
	var firstErr error
	for _, wf := range mc.clients {
		err := wf.SendSpan(name, startMillis, durationMillis, source, traceID, spanID, parents, followsFrom, tags, spanLogs)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (mc *multiClient) SendEvent(name string, startMillis, endMillis int64, source string, tags map[string]string, setters ...event.Option) error {
	var firstErr error
	for _, wf := range mc.clients {
		err := wf.SendEvent(name, startMillis, endMillis, source, tags, setters...)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (mc *multiClient) Flush() error {
	var firstErr error
	for _, wf := range mc.clients {
		err := wf.Flush()
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (mc *multiClient) GetFailureCount() int64 {
	var fc int64
	for _, wf := range mc.clients {
		fc += wf.GetFailureCount()
	}
	return fc
}

func (mc *multiClient) Start() {
	for _, wf := range mc.clients {
		wf.Start()
	}
}

func (mc *multiClient) Close() {
	for _, wf := range mc.clients {
		wf.Close()
	}
}
