package wavefront_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-sdk-go/wavefront"
)

var server http.Server

const (
	wfPort    = "8080"
	proxyPort = "8081"
	token     = "DUMMY_TOKEN"
)

func init() {
	wf := http.Server{Addr: "localhost:" + wfPort}
	proxy := http.Server{Addr: "localhost:" + proxyPort}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.Host, wfPort) {
			if strings.HasSuffix(r.Header.Get("Authorization"), token) {
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		if strings.HasSuffix(r.Host, proxyPort) {
			if len(r.Header.Get("Authorization")) == 0 {
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		w.WriteHeader(http.StatusForbidden)
	})
	go func() { wf.ListenAndServe() }()
	go func() { proxy.ListenAndServe() }()
}

func TestInvalidURL(t *testing.T) {
	_, err := wavefront.NewClient("tut:ut:u")
	assert.NotNil(t, err)
}

func TestDirectSends(t *testing.T) {
	wf, err := wavefront.NewClient("http://" + token + "@localhost:" + wfPort)
	assert.Nil(t, err)
	if wf != nil {
		doTest(t, wf)
	}
}

func TestProxySends(t *testing.T) {
	wf, err := wavefront.NewClient("http://localhost:" + proxyPort)
	assert.Nil(t, err)
	if wf != nil {
		doTest(t, wf)
	}
}

func doTest(t *testing.T, wf wavefront.Client) {
	if err := wf.SendMetric("new-york.power.usage", 42422.0, 0, "go_test", map[string]string{"env": "test"}); err != nil {
		t.Error("Failed SendMetric", err)
	}
	wf.Flush()
	wf.Close()
	assert.Equal(t, int64(0), wf.GetFailureCount(), "GetFailureCount")
	server.Shutdown(context.Background())
}

// func TestDirectSends(t *testing.T) {
// 	directCfg := &DirectConfiguration{
// 		Server:               "http://localhost:" + testPort,
// 		Token:                "DUMMY_TOKEN",
// 		BatchSize:            10000,
// 		MaxBufferSize:        50000,
// 		FlushIntervalSeconds: 1,
// 	}

// 	var err error
// 	if direct, err = NewDirectSender(directCfg); err != nil {
// 		t.Error("Failed Creating Sender", err)
// 	}

// 	if err = direct.SendMetric("new-york.power.usage", 42422.0, 0, "go_test", map[string]string{"env": "test"}); err != nil {
// 		t.Error("Failed SendMetric", err)
// 	}
// 	if err = direct.SendDeltaCounter("lambda.thumbnail.generate", 10.0, "thumbnail_service", map[string]string{"format": "jpeg"}); err != nil {
// 		t.Error("Failed SendDeltaCounter", err)
// 	}

// 	centroids := []histogram.Centroid{
// 		{Value: 30.0, Count: 20},
// 		{Value: 5.1, Count: 10},
// 	}

// 	hgs := map[histogram.Granularity]bool{
// 		histogram.MINUTE: true,
// 		histogram.HOUR:   true,
// 		histogram.DAY:    true,
// 	}

// 	if err = direct.SendDistribution("request.latency", centroids, hgs, 0, "appServer1", map[string]string{"region": "us-west"}); err != nil {
// 		t.Error("Failed SendDistribution", err)
// 	}

// 	if err = direct.SendSpan("getAllUsers", 0, 343500, "localhost",
// 		"7b3bf470-9456-11e8-9eb6-529269fb1459", "0313bafe-9457-11e8-9eb6-529269fb1459",
// 		[]string{"2f64e538-9457-11e8-9eb6-529269fb1459"}, nil,
// 		[]SpanTag{
// 			{Key: "application", Value: "Wavefront"},
// 			{Key: "http.method", Value: "GET"},
// 		},
// 		nil); err != nil {
// 		t.Error("Failed SendSpan", err)
// 	}

// 	direct.Flush()
// 	direct.Close()
// 	if direct.GetFailureCount() > 0 {
// 		t.Error("FailureCount =", direct.GetFailureCount())
// 	}

// 	server.Shutdown(context.Background())
// }
