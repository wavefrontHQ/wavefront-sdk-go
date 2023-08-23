package internal

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wavefronthq/wavefront-sdk-go/internal/auth"
	"net/http"
	"testing"
	"time"
)

func TestReporter_BuildRequest(t *testing.T) {
	var r *reporter
	r = NewReporter("http://localhost:8010/wavefront", auth.NewNoopTokenService(), &http.Client{}).(*reporter)
	request, err := r.buildRequest("wavefront", nil)
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8010/wavefront/report?f=wavefront", request.URL.String())
}

func TestNewClient_WithNilTLSConfig(t *testing.T) {
	client := NewClient(10*time.Second, nil)
	assert.Equal(t, nil, client.Transport)
}

func TestNewClient_WithCustomTLSConfig(t *testing.T) {
	caCertPool := x509.NewCertPool()
	fakeCert := []byte("Not a real cert")
	caCertPool.AppendCertsFromPEM(fakeCert)

	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}

	emptyTLSConfig := &tls.Config{}

	transport := &http.Transport{TLSClientConfig: tlsConfig}
	transportWithEmptyTLSConfig := &http.Transport{TLSClientConfig: emptyTLSConfig}

	client := NewClient(10*time.Second, tlsConfig)
	assert.Equal(t, transport, client.Transport)
	assert.NotEqual(t, transportWithEmptyTLSConfig, client.Transport)
}
