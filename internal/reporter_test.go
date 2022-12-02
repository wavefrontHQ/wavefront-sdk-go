package internal

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestBuildRequest(t *testing.T) {
	var r *reporter
	r = NewReporter("http://localhost:8010/wavefront", "", &http.Client{}).(*reporter)
	request, err := r.buildRequest("wavefront", nil)
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8010/wavefront/report?f=wavefront", request.URL.String())
}
