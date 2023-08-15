package senders

import (
	"crypto/tls"
	"crypto/x509"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvalidURL(t *testing.T) {
	_, err := createConfig("%%%%")
	assert.Error(t, err)
}

func TestScheme(t *testing.T) {
	_, err := createConfig("http://localhost")
	require.NoError(t, err)
	_, err = createConfig("https://localhost")
	require.NoError(t, err)

	_, err = createConfig("gopher://localhost")
	require.Error(t, err)
}

func TestDefaultPortsProxy(t *testing.T) {
	cfg, err := createConfig("http://localhost")
	require.NoError(t, err)
	assert.Equal(t, 2878, cfg.MetricsPort)
	assert.Equal(t, 30001, cfg.TracesPort)
}

func TestMetricPrefixProxy(t *testing.T) {
	cfg, err := createConfig("http://localhost")
	require.NoError(t, err)
	assert.False(t, cfg.Direct())
	assert.Equal(t, "~sdk.go.core.sender.proxy", cfg.MetricPrefix())
}

func TestMetricPrefixDirect(t *testing.T) {
	cfg, err := createConfig("http://11111111-2222-3333-4444-555555555555@localhost")
	require.NoError(t, err)
	assert.True(t, cfg.Direct())
	assert.Equal(t, "~sdk.go.core.sender.direct", cfg.MetricPrefix())
}
func TestDefaultPortsDIHttp(t *testing.T) {
	cfg, err := createConfig("http://11111111-2222-3333-4444-555555555555@localhost")
	require.NoError(t, err)
	assert.Equal(t, 80, cfg.MetricsPort)
	assert.Equal(t, 80, cfg.TracesPort)
}

func TestDefaultPortsDIHttps(t *testing.T) {
	cfg, err := createConfig("https://11111111-2222-3333-4444-555555555555@localhost")
	require.NoError(t, err)
	assert.Equal(t, 443, cfg.MetricsPort)
	assert.Equal(t, 443, cfg.TracesPort)
}

func TestPortExtractedFromURL(t *testing.T) {
	cfg, err := createConfig("http://localhost:1234")
	require.NoError(t, err)
	assert.Equal(t, 1234, cfg.MetricsPort)
	assert.Equal(t, 1234, cfg.TracesPort)
}

func TestUrlWithPortAndPath(t *testing.T) {
	cfg, err := createConfig("http://localhost:8071/wavefront")
	require.NoError(t, err)
	assert.Equal(t, 8071, cfg.MetricsPort)
	assert.Equal(t, 8071, cfg.TracesPort)
	assert.Equal(t, "http://localhost", cfg.Server)
	assert.Equal(t, "/wavefront", cfg.Path)
}

func TestMetricsURLWithPortAndPath(t *testing.T) {
	cfg, err := createConfig("http://localhost:8071/wavefront")
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8071/wavefront", cfg.metricsURL())
	assert.Equal(t, "http://localhost:8071/wavefront", cfg.tracesURL())
}

func TestToken(t *testing.T) {
	cfg, err := createConfig("https://my-api-token@localhost")
	require.NoError(t, err)

	assert.Equal(t, "my-api-token", cfg.Token)
	assert.Equal(t, "https://localhost", cfg.Server)
}

func TestDefaults(t *testing.T) {
	cfg, err := createConfig("https://localhost")
	require.NoError(t, err)

	assert.Equal(t, 10000, cfg.BatchSize)
	assert.Equal(t, 1*time.Second, cfg.FlushInterval)
	assert.Equal(t, 50000, cfg.MaxBufferSize)
	assert.Equal(t, 2878, cfg.MetricsPort)
	assert.Equal(t, 30001, cfg.TracesPort)
	assert.Equal(t, 10*time.Second, cfg.Timeout)
	assert.Equal(t, (*tls.Config)(nil), cfg.TLSConfig)
}

func TestBatchSize(t *testing.T) {
	cfg, err := createConfig("https://localhost", BatchSize(123))
	require.NoError(t, err)

	assert.Equal(t, 123, cfg.BatchSize)
}

func TestFlushIntervalSeconds(t *testing.T) {
	cfg, err := createConfig("https://localhost", FlushIntervalSeconds(123))
	require.NoError(t, err)

	assert.Equal(t, 123*time.Second, cfg.FlushInterval)
}

func TestFlushInterval(t *testing.T) {
	cfg, err := createConfig("https://localhost", FlushInterval(1*time.Hour))
	require.NoError(t, err)

	assert.Equal(t, 1*time.Hour, cfg.FlushInterval)
}

func TestMaxBufferSize(t *testing.T) {
	cfg, err := createConfig("https://localhost", MaxBufferSize(123))
	require.NoError(t, err)

	assert.Equal(t, 123, cfg.MaxBufferSize)
}

func TestMetricsPort(t *testing.T) {
	cfg, err := createConfig("https://localhost", MetricsPort(123))
	require.NoError(t, err)

	assert.Equal(t, 123, cfg.MetricsPort)
}

func TestTracesPort(t *testing.T) {
	cfg, err := createConfig("https://localhost", TracesPort(123))
	require.NoError(t, err)

	assert.Equal(t, 123, cfg.TracesPort)
}

func TestSDKMetricsTags(t *testing.T) {
	cfg, err := createConfig("https://localhost", SDKMetricsTags(map[string]string{"foo": "bar"}), SDKMetricsTags(map[string]string{"foo1": "bar1"}))
	require.NoError(t, err)

	assert.Equal(t, "bar", cfg.SDKMetricsTags["foo"])
	assert.Equal(t, "bar1", cfg.SDKMetricsTags["foo1"])
}

func TestTimeout(t *testing.T) {
	cfg, err := createConfig("https://localhost", Timeout(60*time.Second))
	require.NoError(t, err)

	assert.Equal(t, 60*time.Second, cfg.Timeout)
}

func TestTLSConfigOptions(t *testing.T) {
	caCertPool := x509.NewCertPool()
	fakeCert := []byte("Not a real cert")
	caCertPool.AppendCertsFromPEM(fakeCert)

	tlsConfig := tls.Config{
		RootCAs: caCertPool,
	}
	cfg, err := createConfig("https://localhost", TLSConfigOptions(&tlsConfig))
	require.NoError(t, err)
	assert.Equal(t, caCertPool, cfg.TLSConfig.RootCAs)
}

func TestSDKMetricsTags_Immutability(t *testing.T) {
	map1 := map[string]string{"foo": "bar"}
	map2 := map[string]string{"baz": "none"}
	option1 := SDKMetricsTags(map1)
	option2 := SDKMetricsTags(map2)
	map1["foo"] = "wrong"
	map2["baz"] = "wrong"
	cfg, err := createConfig("https://localhost", option1, option2)
	require.NoError(t, err)
	assert.Equal(t, "bar", cfg.SDKMetricsTags["foo"])
	assert.Equal(t, "none", cfg.SDKMetricsTags["baz"])

	cfg2, err := createConfig("https://localhost", option1)
	require.NoError(t, err)
	assert.Equal(t, "bar", cfg2.SDKMetricsTags["foo"])
	_, ok := cfg2.SDKMetricsTags["baz"]
	assert.False(t, ok)
}
