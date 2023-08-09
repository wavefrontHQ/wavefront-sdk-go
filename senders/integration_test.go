package senders

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestEndToEnd(t *testing.T) {
	testServer := startTestServer()
	defer testServer.Close()
	sender, err := NewSender(testServer.URL)
	require.NoError(t, err)
	require.NoError(t, sender.SendMetric("my metric", 20, 0, "localhost", nil))
	require.NoError(t, sender.Flush())

	assert.Equal(t, 1, len(testServer.MetricLines))
	assert.Equal(t, "\"my-metric\" 20 source=\"localhost\"", testServer.MetricLines[0])
	assert.Equal(t, "/report?f=wavefront", testServer.LastRequestURL)
}

func TestEndToEndWithPath(t *testing.T) {
	testServer := startTestServer()
	defer testServer.Close()
	sender, err := NewSender(testServer.URL + "/test-path")
	require.NoError(t, err)
	require.NoError(t, sender.SendMetric("my metric", 20, 0, "localhost", nil))
	require.NoError(t, sender.Flush())

	assert.Equal(t, 1, len(testServer.MetricLines))
	assert.Equal(t, "\"my-metric\" 20 source=\"localhost\"", testServer.MetricLines[0])
	assert.Equal(t, "/test-path/report?f=wavefront", testServer.LastRequestURL)
}

func TestTLSEndToEnd(t *testing.T) {
	testServer := startTLSTestServer()
	defer testServer.Close()
	testServer.httpServer.Client()
	tlsConfig := testServer.TLSConfig()

	sender, err := NewSender(testServer.URL, TLSConfigOptions(tlsConfig))
	require.NoError(t, err)
	require.NoError(t, sender.SendMetric("my metric", 20, 0, "localhost", nil))
	require.NoError(t, sender.Flush())

	assert.Equal(t, 1, len(testServer.MetricLines))
	assert.Equal(t, "\"my-metric\" 20 source=\"localhost\"", testServer.MetricLines[0])
}

func TestEndToEndWithInternalMetrics(t *testing.T) {
	waitTill1stInternalMetricsCollection := time.Millisecond * 1500
	internalMetricsTicker := time.NewTicker(time.Millisecond * 900)

	testServer := startTestServer()
	defer testServer.Close()

	sender, err := NewSender(testServer.URL, InternalMetricsTicker(internalMetricsTicker))
	require.NoError(t, err)
	require.NoError(t, sender.SendMetric("my metric", 20, 0, "localhost", nil))
	time.Sleep(waitTill1stInternalMetricsCollection)
	require.NoError(t, sender.Flush())
	metricLines := testServer.MetricLines

	assert.Equal(t, true, testServer.hasReceivedInternalMetric("points.valid"))
	assert.Equal(t, 12, len(metricLines))
	assert.Equal(t, "\"my-metric\" 20 source=\"localhost\"", metricLines[0])
	assert.Equal(t, "/report?f=wavefront", testServer.LastRequestURL)
}

func TestEndToEndWithInternalMetricsDisabled(t *testing.T) {
	waitTill1stInternalMetricsCollection := time.Millisecond * 1500
	internalMetricsTicker := time.NewTicker(time.Millisecond * 900)

	testServer := startTestServer()
	defer testServer.Close()

	sender, err := NewSender(testServer.URL, InternalMetricsTicker(internalMetricsTicker), InternalMetricsEnabled(false))
	require.NoError(t, err)
	require.NoError(t, sender.SendMetric("my metric", 20, 0, "localhost", nil))
	time.Sleep(waitTill1stInternalMetricsCollection)
	require.NoError(t, sender.Flush())
	metricLines := testServer.MetricLines

	assert.Equal(t, false, testServer.hasReceivedInternalMetric("points.valid"))
	assert.Equal(t, 1, len(metricLines))
	assert.Equal(t, "\"my-metric\" 20 source=\"localhost\"", metricLines[0])
	assert.Equal(t, "/report?f=wavefront", testServer.LastRequestURL)
}
