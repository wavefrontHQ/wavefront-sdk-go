package senders

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEndToEnd(t *testing.T) {
	testServer := startTestServer(false)
	defer testServer.Close()
	sender, err := NewSender(testServer.URL)
	require.NoError(t, err)
	require.NoError(t, sender.SendMetric("my metric", 20, 0, "localhost", nil))
	require.NoError(t, sender.SendEvent("dramatic event", 20, 0, "localhost", nil))
	require.NoError(t, sender.Flush())

	assert.Equal(t, 1, len(testServer.MetricLines))
	assert.Equal(t, "\"my-metric\" 20 source=\"localhost\"", testServer.MetricLines[0])
	assert.Equal(t, "@Event 20000 20001 \"dramatic event\" host=\"localhost\"", testServer.EventLines[0])
	assert.Equal(t, "/report?f=wavefront", testServer.RequestURLs[0])
	assert.Equal(t, "/api/v2/event", testServer.RequestURLs[1])
}

func TestEndToEndWithPath(t *testing.T) {
	testServer := startTestServer(false)
	defer testServer.Close()
	sender, err := NewSender(testServer.URL + "/test-path")
	require.NoError(t, err)
	require.NoError(t, sender.SendMetric("my metric", 20, 0, "localhost", nil))
	require.NoError(t, sender.Flush())

	assert.Equal(t, 1, len(testServer.MetricLines))
	assert.Equal(t, "\"my-metric\" 20 source=\"localhost\"", testServer.MetricLines[0])
	assert.Equal(t, "/test-path/report?f=wavefront", testServer.RequestURLs[0])
}

func TestTLSEndToEnd(t *testing.T) {
	testServer := startTestServer(true)
	defer testServer.Close()
	testServer.httpServer.Client()
	tlsConfig := testServer.TLSConfig()

	sender, err := NewSender(testServer.URL, TLSConfigOptions(tlsConfig), Timeout(10*time.Second))
	require.NoError(t, err)
	require.NoError(t, sender.SendMetric("my metric", 20, 0, "localhost", nil))
	require.NoError(t, sender.Flush())

	assert.Equal(t, 1, len(testServer.MetricLines))
	assert.Equal(t, "\"my-metric\" 20 source=\"localhost\"", testServer.MetricLines[0])
}

func TestEndToEndWithInternalMetrics(t *testing.T) {
	testServer := startTestServer(false)
	defer testServer.Close()

	sender, err := NewSender(testServer.URL, SendInternalMetrics(true))
	require.NoError(t, err)
	require.NoError(t, sender.SendMetric("my metric", 20, 0, "localhost", nil))
	sender.(*realSender).internalRegistry.Flush()
	require.NoError(t, sender.Flush())
	metricLines := testServer.MetricLines

	assert.Equal(t, true, testServer.hasReceivedLine("points.valid"))
	assert.Equal(t, 12, len(metricLines))
	assert.Equal(t, "\"my-metric\" 20 source=\"localhost\"", metricLines[0])
	assert.Equal(t, "/report?f=wavefront", testServer.RequestURLs[0])
}

func TestEndToEndWithoutInternalMetrics(t *testing.T) {
	testServer := startTestServer(false)
	defer testServer.Close()

	sender, err := NewSender(testServer.URL, SendInternalMetrics(false))
	require.NoError(t, err)
	require.NoError(t, sender.SendMetric("my metric", 20, 0, "localhost", nil))
	sender.(*realSender).internalRegistry.Flush()
	require.NoError(t, sender.Flush())
	metricLines := testServer.MetricLines

	assert.Equal(t, false, testServer.hasReceivedLine("points.valid"))
	assert.Equal(t, 1, len(metricLines))
	assert.Equal(t, "\"my-metric\" 20 source=\"localhost\"", metricLines[0])
	assert.Equal(t, "/report?f=wavefront", testServer.RequestURLs[0])
}

func TestEndToEndDirect(t *testing.T) {
	testServer := startTestServer(false)
	defer testServer.Close()
	sender, err := NewSender(testServer.URL, APIToken("just make it look direct"))
	require.NoError(t, err)
	require.NoError(t, sender.SendMetric("my metric", 20, 0, "localhost", nil))
	require.NoError(t, sender.SendEvent("dramatic event", 20, 0, "localhost", nil))
	require.NoError(t, sender.Flush())

	assert.Equal(t, 1, len(testServer.MetricLines))
	assert.Equal(t, "\"my-metric\" 20 source=\"localhost\"", testServer.MetricLines[0])
	assert.Equal(t, `{"annotations":{},"endTime":20001,"hosts":["localhost"],"name":"dramatic event","startTime":20000}`, testServer.EventLines[0])
	assert.Equal(t, "/report?f=wavefront", testServer.RequestURLs[0])
	assert.Equal(t, "/api/v2/event", testServer.RequestURLs[1])
}
