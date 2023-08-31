package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-sdk-go/internal/auth/csp"
)

func TestCSPService_MultipleCSPRequests(t *testing.T) {
	cspServer := httptest.NewServer(csp.FakeCSPHandler(nil))
	defer cspServer.Close()
	tokenService := NewCSPServerToServerService(cspServer.URL, "a", "b", nil)

	cspTokenService := tokenService.(*CSPService)
	cspTokenService.defaultRefreshInterval = 1 * time.Second

	assert.NotNil(t, tokenService)
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	assert.NoError(t, tokenService.Authorize(req))
	token := req.Header.Get("Authorization")
	assert.NotNil(t, token)
	assert.NotEmpty(t, token)
	assert.NotEqual(t, "INVALID_TOKEN", token)
	assert.Equal(t, "Bearer abc", token)

	time.Sleep(2 * time.Second)
	req, _ = http.NewRequest("GET", "https://example.com", nil)
	assert.NoError(t, tokenService.Authorize(req))
	token = req.Header.Get("Authorization")

	assert.NotNil(t, token)
	assert.NotEmpty(t, token)
	assert.NotEqual(t, "INVALID_TOKEN", token)
	assert.Equal(t, "Bearer def", token)
	tokenService.Close()
}

func TestCSPService_WhenAuthenticationFails_AuthorizeReturnsError(t *testing.T) {
	cspServer := httptest.NewServer(csp.FakeCSPHandler(nil))
	defer cspServer.Close()
	tokenService := NewCSPServerToServerService(cspServer.URL, "nope", "wrong", nil)
	defer tokenService.Close()

	cspTokenService := tokenService.(*CSPService)
	cspTokenService.defaultRefreshInterval = 1 * time.Second

	assert.NotNil(t, tokenService)
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	assert.Error(t, tokenService.Authorize(req))
	token := req.Header.Get("Authorization")
	assert.Equal(t, "", token)
}
