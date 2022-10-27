package senders_test

import (
	"github.com/wavefronthq/wavefront-sdk-go/internal/spans"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
	"github.com/wavefronthq/wavefront-sdk-go/senders"
)

func TestNoOpSender(t *testing.T) {
	wf, err := senders.NewSender("http://DUMMY_TOKEN@localhost:8080")
	assert.Nil(t, err)
	if wf == nil {
		t.Error("Failed Initialize NewSender", err)
	}

	if err := wf.SendMetric("new-york.power.usage", 42422.0, 0, "go_test", map[string]string{"env": "test"}); err != nil {
		t.Error("Failed SendMetric", err)
	}

	centroids := []histogram.Centroid{
		{Value: 30.0, Count: 20},
		{Value: 5.1, Count: 10},
	}

	hgs := map[histogram.Granularity]bool{
		histogram.MINUTE: true,
		histogram.HOUR:   true,
		histogram.DAY:    true,
	}

	if err := wf.SendDistribution("request.latency", centroids, hgs, 0, "appServer1", map[string]string{"region": "us-west"}); err != nil {
		t.Error("Failed SendDistribution", err)
	}

	if err := wf.SendDeltaCounter("invocation.count", 0, "appServer2", map[string]string{"region": "us-west"}); err != nil {
		t.Error("Failed SendDeltaCounter", err)
	}

	if err := wf.SendSpan("getAllUsers", 0, 343500, "localhost",
		"7b3bf470-9456-11e8-9eb6-529269fb1459", "0313bafe-9457-11e8-9eb6-529269fb1459",
		[]string{"2f64e538-9457-11e8-9eb6-529269fb1459"}, nil,
		[]spans.SpanTag{
			{Key: "application", Value: "Wavefront"},
			{Key: "http.method", Value: "GET"},
		},
		nil); err != nil {
		t.Error("Failed SendSpan", err)
	}

	if err := wf.SendEvent("updateAllUsers", 0, 37484, "localhost", map[string]string{"region": "us-west"}); err != nil {
		t.Error("Failed SendEvent", err)
	}

	wf.Flush()
	wf.Close()
	assert.Equal(t, int64(0), wf.GetFailureCount(), "GetFailureCount")
}
