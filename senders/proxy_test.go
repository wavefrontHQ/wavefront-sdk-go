package senders_test

import (
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
	"github.com/wavefronthq/wavefront-sdk-go/senders"
	"io"
	"net"
	"os"
	"testing"
	"time"
)

func netcat(addr string, keepopen bool, portCh chan int) {
	laddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		panic("netcat resolve " + addr + " failed with: " + err.Error())
	}

	lis, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		panic("netcat listen " + addr + " failed with: " + err.Error())
	}
	portCh <- lis.Addr().(*net.TCPAddr).Port

	for loop := true; loop; loop = keepopen {
		conn, err := lis.Accept()
		if err != nil {
			panic("netcat accept " + addr + " failed with: " + err.Error())
		}
		_, err = io.Copy(os.Stdout, conn)
		if err != nil {
			panic("netcat copy " + addr + " failed with: " + err.Error())
		}
	}
	err = lis.Close()
	if err != nil {
		panic("netcat close " + addr + " failed with: " + err.Error())
	}

}

func TestProxySends(t *testing.T) {

	ports := getConnection(t)

	proxyCfg := &senders.ProxyConfiguration{
		Host:                 "localhost",
		MetricsPort:          ports[0],
		DistributionPort:     ports[1],
		TracingPort:          ports[2],
		FlushIntervalSeconds: 10,
	}

	var err error
	var proxy senders.Sender
	if proxy, err = senders.NewProxySender(proxyCfg); err != nil {
		t.Error("Failed Creating Sender", err)
	}

	verifyResults(t, err, proxy)

	proxy.Flush()
	proxy.Close()
	if proxy.GetFailureCount() > 0 {
		t.Error("FailureCount =", proxy.GetFailureCount())
	}
}

func getConnection(t *testing.T) [3]int {
	ch := make(chan int, 3)
	var ports [3]int

	go netcat("localhost:0", false, ch)
	go netcat("localhost:0", false, ch)
	go netcat("localhost:0", false, ch)

	for i := 0; i < 3; {
		select {
		case ports[i] = <-ch:
			i++
		case <-time.After(time.Second):
			t.Fail()
			t.Logf("Could not get netcats")
		}
	}

	return ports
}

func TestProxySendsWithTags(t *testing.T) {

	ports := getConnection(t)

	proxyCfg := &senders.ProxyConfiguration{
		Host:                 "localhost",
		MetricsPort:          ports[0],
		DistributionPort:     ports[1],
		TracingPort:          ports[2],
		FlushIntervalSeconds: 10,
		SDKMetricsTags:       map[string]string{"foo": "bar"},
	}

	var err error
	var proxy senders.Sender
	if proxy, err = senders.NewProxySender(proxyCfg); err != nil {
		t.Error("Failed Creating Sender", err)
	}

	verifyResults(t, err, proxy)

	proxy.Flush()
	proxy.Close()
	if proxy.GetFailureCount() > 0 {
		t.Error("FailureCount =", proxy.GetFailureCount())
	}
}

func verifyResults(t *testing.T, err error, proxy senders.Sender) {
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
}
