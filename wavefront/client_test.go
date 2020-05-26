package wavefront_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-sdk-go/senders"
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

func TestError(t *testing.T) {
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

func doTest(t *testing.T, wf senders.Sender) {
	if err := wf.SendMetric("new-york.power.usage", 42422.0, 0, "go_test", map[string]string{"env": "test"}); err != nil {
		t.Error("Failed SendMetric", err)
	}
	wf.Flush()
	wf.Close()
	assert.Equal(t, int64(0), wf.GetFailureCount(), "GetFailureCount")
	server.Shutdown(context.Background())
}
