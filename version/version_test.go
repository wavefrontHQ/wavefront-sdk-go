package version

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanForVersion(t *testing.T) {
	bi := &debug.BuildInfo{
		Deps: []*debug.Module{
			{
				Path:    "not_right_module",
				Version: "v1.0.2",
			},
			{
				Path:    "github.com/wavefronthq/wavefront-sdk-go",
				Version: "v0.10.3",
			},
		},
	}
	assert.Equal(t, "0.10.3", scanForVersion(bi, true))
}

func TestScanForVersionMissing(t *testing.T) {
	bi := &debug.BuildInfo{
		Deps: []*debug.Module{
			{
				Path:    "not_right_module",
				Version: "v1.0.2",
			},
		},
	}
	assert.Equal(t, "unavailable", scanForVersion(bi, true))
}

func TestScanForVersionNone(t *testing.T) {
	assert.Equal(t, "unavailable", scanForVersion(nil, false))
}
