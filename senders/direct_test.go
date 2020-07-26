package senders_test

import (
	"testing"

	"github.com/wavefronthq/wavefront-sdk-go/histogram"
	"github.com/wavefronthq/wavefront-sdk-go/senders"
)

var direct senders.Sender

func TestDirectSends(t *testing.T) {
	directCfg := &senders.DirectConfiguration{
		Server:               "http://localhost:" + wfPort,
		Token:                "DUMMY_TOKEN",
		BatchSize:            10000,
		MaxBufferSize:        500000,
		FlushIntervalSeconds: 1,
	}

	var err error
	if direct, err = senders.NewDirectSender(directCfg); err != nil {
		t.Error("Failed Creating Sender", err)
	}

	if err = direct.SendMetric("new-york.power.usage", 42422.0, 0, "go_test", map[string]string{"env": "test"}); err != nil {
		t.Error("Failed SendMetric", err)
	}
	if err = direct.SendDeltaCounter("lambda.thumbnail.generate", 10.0, "thumbnail_service", map[string]string{"format": "jpeg"}); err != nil {
		t.Error("Failed SendDeltaCounter", err)
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

	if err = direct.SendDistribution("request.latency", centroids, hgs, 0, "appServer1", map[string]string{"region": "us-west"}); err != nil {
		t.Error("Failed SendDistribution", err)
	}

	if err = direct.SendSpan("getAllUsers", 0, 343500, "localhost",
		"7b3bf470-9456-11e8-9eb6-529269fb1459", "0313bafe-9457-11e8-9eb6-529269fb1459",
		[]string{"2f64e538-9457-11e8-9eb6-529269fb1459"}, nil,
		[]senders.SpanTag{
			{Key: "application", Value: "Wavefront"},
			{Key: "http.method", Value: "GET"},
		},
		nil); err != nil {
		t.Error("Failed SendSpan", err)
	}

	direct.Flush()
	direct.Close()
	if direct.GetFailureCount() > 0 {
		t.Error("FailureCount =", direct.GetFailureCount())
	}

}
