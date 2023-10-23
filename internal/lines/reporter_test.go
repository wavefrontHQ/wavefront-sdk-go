package lines

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wavefronthq/wavefront-sdk-go/internal/auth"
)

func TestReporter_BuildRequest(t *testing.T) {
	r := NewReporter("http://localhost:8010/wavefront", auth.NewNoopTokenService(), &http.Client{}).(*reporter)
	request, err := r.buildRequest("wavefront", nil)
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8010/wavefront/report?f=wavefront", request.URL.String())
}
