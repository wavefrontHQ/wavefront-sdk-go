package senders_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wavefronthq/wavefront-sdk-go/senders"
)

func TestInvalidURL(t *testing.T) {
	_, err := senders.CreateConfig("%%%%")
	assert.Error(t, err)
}

func TestScheme(t *testing.T) {
	_, err := senders.CreateConfig("http://localhost")
	require.NoError(t, err)
	_, err = senders.CreateConfig("https://localhost")
	require.NoError(t, err)

	_, err = senders.CreateConfig("gopher://localhost")
	require.Error(t, err)
}

func TestDefaultPortsProxy(t *testing.T) {
	cfg, err := senders.CreateConfig("http://localhost")
	require.NoError(t, err)
	assert.Equal(t, 2878, cfg.MetricsPort)
	assert.Equal(t, 30001, cfg.TracesPort)
}

func TestMetricPrefixProxy(t *testing.T) {
	cfg, err := senders.CreateConfig("http://localhost")
	require.NoError(t, err)
	assert.False(t, cfg.Direct())
	assert.Equal(t, "~sdk.go.core.sender.proxy", cfg.MetricPrefix())
}

func TestMetricPrefixDirect(t *testing.T) {
	cfg, err := senders.CreateConfig("http://11111111-2222-3333-4444-555555555555@localhost")
	require.NoError(t, err)
	assert.True(t, cfg.Direct())
	assert.Equal(t, "~sdk.go.core.sender.direct", cfg.MetricPrefix())
}
func TestDefaultPortsDIHttp(t *testing.T) {
	cfg, err := senders.CreateConfig("http://11111111-2222-3333-4444-555555555555@localhost")
	require.NoError(t, err)
	assert.Equal(t, 80, cfg.MetricsPort)
	assert.Equal(t, 80, cfg.TracesPort)
}

func TestDefaultPortsDIHttps(t *testing.T) {
	cfg, err := senders.CreateConfig("https://11111111-2222-3333-4444-555555555555@localhost")
	require.NoError(t, err)
	assert.Equal(t, 443, cfg.MetricsPort)
	assert.Equal(t, 443, cfg.TracesPort)
}

func TestPortExtractedFromURL(t *testing.T) {
	cfg, err := senders.CreateConfig("http://localhost:1234")
	require.NoError(t, err)
	assert.Equal(t, 1234, cfg.MetricsPort)
	assert.Equal(t, 1234, cfg.TracesPort)
}

func TestToken(t *testing.T) {
	cfg, err := senders.CreateConfig("https://my-api-token@localhost")
	require.NoError(t, err)

	assert.Equal(t, "my-api-token", cfg.Token)
	assert.Equal(t, "https://localhost", cfg.Server)
}

func TestDefaults(t *testing.T) {
	cfg, err := senders.CreateConfig("https://localhost")
	require.NoError(t, err)

	assert.Equal(t, 10000, cfg.BatchSize)
	assert.Equal(t, 1, cfg.FlushIntervalSeconds)
	assert.Equal(t, 50000, cfg.MaxBufferSize)
	assert.Equal(t, 2878, cfg.MetricsPort)
	assert.Equal(t, 30001, cfg.TracesPort)
}

func TestBatchSize(t *testing.T) {
	cfg, err := senders.CreateConfig("https://localhost", senders.BatchSize(123))
	require.NoError(t, err)

	assert.Equal(t, 123, cfg.BatchSize)
}

func TestFlushIntervalSeconds(t *testing.T) {
	cfg, err := senders.CreateConfig("https://localhost", senders.FlushIntervalSeconds(123))
	require.NoError(t, err)

	assert.Equal(t, 123, cfg.FlushIntervalSeconds)
}

func TestMaxBufferSize(t *testing.T) {
	cfg, err := senders.CreateConfig("https://localhost", senders.MaxBufferSize(123))
	require.NoError(t, err)

	assert.Equal(t, 123, cfg.MaxBufferSize)
}

func TestMetricsPort(t *testing.T) {
	cfg, err := senders.CreateConfig("https://localhost", senders.MetricsPort(123))
	require.NoError(t, err)

	assert.Equal(t, 123, cfg.MetricsPort)
}

func TestTracesPort(t *testing.T) {
	cfg, err := senders.CreateConfig("https://localhost", senders.TracesPort(123))
	require.NoError(t, err)

	assert.Equal(t, 123, cfg.TracesPort)
}

func TestSDKMetricsTags(t *testing.T) {
	cfg, err := senders.CreateConfig("https://localhost", senders.SDKMetricsTags(map[string]string{"foo": "bar"}), senders.SDKMetricsTags(map[string]string{"foo1": "bar1"}))
	require.NoError(t, err)

	assert.Equal(t, "bar", cfg.SDKMetricsTags["foo"])
	assert.Equal(t, "bar1", cfg.SDKMetricsTags["foo1"])
}

func TestSDKMetricsTags_Immutability(t *testing.T) {
	map1 := map[string]string{"foo": "bar"}
	map2 := map[string]string{"baz": "none"}
	option1 := senders.SDKMetricsTags(map1)
	option2 := senders.SDKMetricsTags(map2)
	map1["foo"] = "wrong"
	map2["baz"] = "wrong"
	cfg, err := senders.CreateConfig("https://localhost", option1, option2)
	require.NoError(t, err)
	assert.Equal(t, "bar", cfg.SDKMetricsTags["foo"])
	assert.Equal(t, "none", cfg.SDKMetricsTags["baz"])

	cfg2, err := senders.CreateConfig("https://localhost", option1)
	require.NoError(t, err)
	assert.Len(t, cfg2.SDKMetricsTags, 1)
}
