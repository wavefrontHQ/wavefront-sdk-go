package senders

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
	"github.com/wavefronthq/wavefront-sdk-go/internal/auth/csp"
)

func TestSendDirect(t *testing.T) {
	token := "direct-send-api-token"
	directServer := startTestServer(false)
	defer directServer.Close()
	updatedUrl, err := url.Parse(directServer.URL)
	updatedUrl.User = url.User(token)
	wf, err := NewSender(updatedUrl.String())

	require.NoError(t, err)
	testSender(t, wf, directServer)
	assert.Equal(t,
		[]string{
			"Bearer direct-send-api-token",
			"Bearer direct-send-api-token",
			"Bearer direct-send-api-token",
		},
		directServer.AuthHeaders)
}

func TestSendDirectWithTags(t *testing.T) {
	token := "direct-send-api-token"
	directServer := startTestServer(false)
	defer directServer.Close()

	updatedUrl, err := url.Parse(directServer.URL)
	updatedUrl.User = url.User(token)
	tags := map[string]string{"foo": "bar"}
	wf, err := NewSender(updatedUrl.String(), SDKMetricsTags(tags))
	require.NoError(t, err)
	testSender(t, wf, directServer)

	assert.Equal(t,
		[]string{
			"Bearer direct-send-api-token",
			"Bearer direct-send-api-token",
			"Bearer direct-send-api-token",
		},
		directServer.AuthHeaders)
}

func TestSendProxy(t *testing.T) {
	proxyServer := startTestServer(false)
	defer proxyServer.Close()

	wf, err := NewSender(proxyServer.URL)
	require.NoError(t, err)
	testSender(t, wf, proxyServer)
	assert.Equal(t, []string{"", "", ""}, proxyServer.AuthHeaders)
}

func TestSendCSPClientCredentials(t *testing.T) {
	proxyServer := startTestServer(false)
	cspServer := httptest.NewServer(csp.FakeCSPHandler(nil))
	defer proxyServer.Close()
	defer cspServer.Close()

	wf, err := NewSender(proxyServer.URL, CSPClientCredentials(
		"a",
		"b",
		CSPBaseURL(cspServer.URL),
	))
	require.NoError(t, err)
	testSender(t, wf, proxyServer)
	assert.Equal(t, []string{"Bearer abc", "Bearer abc", "Bearer abc"}, proxyServer.AuthHeaders)
}

func TestSendCSPAPIToken(t *testing.T) {
	wavefrontServer := startTestServer(false)
	cspServer := httptest.NewServer(csp.FakeCSPHandler([]string{"12345"}))
	defer wavefrontServer.Close()
	defer cspServer.Close()

	wf, err := NewSender(wavefrontServer.URL, CSPAPIToken(
		"12345",
		CSPBaseURL(cspServer.URL),
	))
	require.NoError(t, err)
	testSender(t, wf, wavefrontServer)
	assert.Equal(t, []string{"Bearer abc", "Bearer abc", "Bearer abc"}, wavefrontServer.AuthHeaders)
}

func testSender(t *testing.T, wf Sender, server *testServer) {
	assert.NoError(t, wf.SendMetric(
		"new-york.power.usage",
		42422.0,
		0,
		"go_test",
		map[string]string{"env": "test"},
	))

	centroids := []histogram.Centroid{
		{Value: 30.0, Count: 20},
		{Value: 5.1, Count: 10},
	}

	hgs := map[histogram.Granularity]bool{
		histogram.MINUTE: true,
		histogram.HOUR:   true,
		histogram.DAY:    true,
	}

	assert.NoError(t, wf.SendDistribution(
		"request.latency",
		centroids,
		hgs,
		0,
		"appServer1",
		map[string]string{"region": "us-west"},
	))

	assert.NoError(t, wf.SendSpan("getAllUsers", 0, 343500, "localhost",
		"7b3bf470-9456-11e8-9eb6-529269fb1459", "0313bafe-9457-11e8-9eb6-529269fb1459",
		[]string{"2f64e538-9457-11e8-9eb6-529269fb1459"}, nil,
		[]SpanTag{
			{Key: "application", Value: "Wavefront"},
			{Key: "http.method", Value: "GET"},
		},
		nil))

	assert.NoError(t, wf.Flush())
	wf.Close()
	assert.Equal(t, int64(0), wf.GetFailureCount(), "GetFailureCount")
	assert.True(t, server.hasReceivedLine("new-york.power.usage"))
	assert.True(t, server.hasReceivedLine("request.latency"))
	assert.True(t, server.hasReceivedLine("0313bafe-9457-11e8-9eb6-529269fb1459"))
}
