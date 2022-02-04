package senders_test

import (
	"io"
	"net"
	"os"
	"testing"

	"github.com/wavefronthq/wavefront-sdk-go/histogram"
	"github.com/wavefronthq/wavefront-sdk-go/senders"
)

var proxy senders.Sender

func netcat(addr string, keepopen bool) {
	laddr, _ := net.ResolveTCPAddr("tcp", addr)
	lis, _ := net.ListenTCP("tcp", laddr)
	for loop := true; loop; loop = keepopen {
		conn, _ := lis.Accept()
		io.Copy(os.Stdout, conn)
	}
	lis.Close()
}

func init() {
	go netcat("localhost:30000", false)
	go netcat("localhost:40000", false)
	go netcat("localhost:50000", false)
}

func TestProxySends(t *testing.T) {
	proxyCfg := &senders.ProxyConfiguration{
		Host:                 "localhost",
		MetricsPort:          30000,
		DistributionPort:     40000,
		TracingPort:          50000,
		FlushIntervalSeconds: 10,
	}

	var err error
	if proxy, err = senders.NewProxySender(proxyCfg); err != nil {
		t.Error("Failed Creating Sender", err)
	}

	if err = proxy.SendMetric("new-york.power.usage", 42422.0, 0, "go_test", map[string]string{"env": "test"}); err != nil {
		t.Error("Failed SendMetric", err)
	}
	if err = proxy.SendDeltaCounter("lambda.thumbnail.generate", 10.0, "thumbnail_service", map[string]string{"format": "jpeg"}); err != nil {
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

	if err = proxy.SendDistribution("request.latency", centroids, hgs, 0, "appServer1", map[string]string{"region": "us-west"}); err != nil {
		t.Error("Failed SendDistribution", err)
	}

	if err = proxy.SendSpan("getAllUsers", 0, 343500, "localhost",
		"7b3bf470-9456-11e8-9eb6-529269fb1459", "0313bafe-9457-11e8-9eb6-529269fb1459",
		[]string{"2f64e538-9457-11e8-9eb6-529269fb1459"}, nil,
		[]senders.SpanTag{
			{Key: "application", Value: "Wavefront"},
			{Key: "http.method", Value: "GET"},
		},
		nil); err != nil {
		t.Error("Failed SendSpan", err)
	}

	proxy.Flush()
	proxy.Close()
	if proxy.GetFailureCount() > 0 {
		t.Error("FailureCount =", proxy.GetFailureCount())
	}
}

func TestProxySendsWithTag(t *testing.T) {
	proxyCfg := &senders.ProxyConfiguration{
		Host:                 "localhost",
		MetricsPort:          30000,
		DistributionPort:     40000,
		TracingPort:          50000,
		FlushIntervalSeconds: 10,
	}

	setter := senders.SetTag("foo", "bar")
	var err error
	if proxy, err = senders.NewProxySender(proxyCfg, setter); err != nil {
		t.Error("Failed Creating Sender", err)
	}

	if err = proxy.SendMetric("new-york.power.usage", 42422.0, 0, "go_test", map[string]string{"env": "test"}); err != nil {
		t.Error("Failed SendMetric", err)
	}
	if err = proxy.SendDeltaCounter("lambda.thumbnail.generate", 10.0, "thumbnail_service", map[string]string{"format": "jpeg"}); err != nil {
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

	if err = proxy.SendDistribution("request.latency", centroids, hgs, 0, "appServer1", map[string]string{"region": "us-west"}); err != nil {
		t.Error("Failed SendDistribution", err)
	}

	if err = proxy.SendSpan("getAllUsers", 0, 343500, "localhost",
		"7b3bf470-9456-11e8-9eb6-529269fb1459", "0313bafe-9457-11e8-9eb6-529269fb1459",
		[]string{"2f64e538-9457-11e8-9eb6-529269fb1459"}, nil,
		[]senders.SpanTag{
			{Key: "application", Value: "Wavefront"},
			{Key: "http.method", Value: "GET"},
		},
		nil); err != nil {
		t.Error("Failed SendSpan", err)
	}

	proxy.Flush()
	proxy.Close()
	if proxy.GetFailureCount() > 0 {
		t.Error("FailureCount =", proxy.GetFailureCount())
	}
}
