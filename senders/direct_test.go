package senders

import (
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/wavefronthq/wavefront-sdk-go/histogram"
)

var direct Sender
var server http.Server

const (
	testPort = "8080"
)

func init() {
	server = http.Server{
		Addr: "localhost:" + testPort,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("\n=========", r.URL, r.Method, r.Header.Get("Authorization"))
		gr, _ := gzip.NewReader(r.Body)
		msg, _ := ioutil.ReadAll(gr)
		fmt.Println(string(msg))
		w.WriteHeader(http.StatusOK)
	})
	go func() {
		server.ListenAndServe()
	}()
}

func TestDirectSends(t *testing.T) {
	directCfg := &DirectConfiguration{
		Server:               "http://localhost:" + testPort,
		Token:                "DUMMY_TOKEN",
		BatchSize:            10000,
		MaxBufferSize:        50000,
		FlushIntervalSeconds: 1,
	}

	var err error
	if direct, err = NewDirectSender(directCfg); err != nil {
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
		[]SpanTag{
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

	server.Shutdown(context.Background())
}
