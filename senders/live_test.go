package senders

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func skipUnlessVarsAreSet(t *testing.T) {
	if os.Getenv("LIVE_TEST_HOST") == "" {
		t.Skip()
	}
}

func TestCSP_LIVE(t *testing.T) {
	skipUnlessVarsAreSet(t)

	sender, err := NewSender(os.Getenv("LIVE_TEST_HOST"),
		CSPClientCredentials(
			os.Getenv("LIVE_TEST_CSP_CLIENT_ID"),
			os.Getenv("LIVE_TEST_CSP_CLIENT_SECRET"),
			CSPBaseURL(os.Getenv("LIVE_TEST_CSP_BASE_URL")),
		))
	assert.NoError(t, err)
	assert.NoError(t, sender.SendMetric("test.go-metrics.can-send", 1, 0, "go test",
		map[string]string{"scenario": "direct-csp-server-to-server"}))
	assert.NoError(t, sender.Flush())
	sender.Close()
}

func TestCSP_API_TOKEN_LIVE(t *testing.T) {
	skipUnlessVarsAreSet(t)

	sender, err := NewSender(os.Getenv("LIVE_TEST_HOST"),
		CSPAPIToken(
			os.Getenv("LIVE_TEST_CSP_API_TOKEN"),
			CSPBaseURL(os.Getenv("LIVE_TEST_CSP_BASE_URL")),
		))
	assert.NoError(t, err)
	assert.NoError(t, sender.SendMetric("test.go-metrics.can-send", 1, 0, "go test",
		map[string]string{"scenario": "direct-csp-api-token"}))
	assert.NoError(t, sender.Flush())
	sender.Close()
}

func TestWF_API_TOKEN_LIVE(t *testing.T) {
	skipUnlessVarsAreSet(t)

	sender, err := NewSender(
		os.Getenv("LIVE_TEST_HOST"),
		APIToken(os.Getenv("LIVE_TEST_WF_API_TOKEN")),
	)
	assert.NoError(t, err)
	assert.NoError(t, sender.SendMetric("test.go-metrics.can-send", 1, 0, "go test",
		map[string]string{"scenario": "direct-wf-token"}))
	assert.NoError(t, sender.Flush())
	sender.Close()
}
