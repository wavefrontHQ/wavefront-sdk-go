package delta

import (
	"fmt"

	"github.com/wavefronthq/wavefront-sdk-go/internal"
	"github.com/wavefronthq/wavefront-sdk-go/internal/metric"
)

func Line(name string, value float64, source string, tags map[string]string, defaultSource string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty metric name")
	}
	if !internal.HasDeltaPrefix(name) {
		name = internal.DeltaCounterName(name)
	}
	return metric.Line(name, value, 0, source, tags, defaultSource)
}
