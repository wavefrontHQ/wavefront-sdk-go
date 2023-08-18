package sdkmetrics

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRealMetricRegistry(t *testing.T) {
	sender := &mockSender{}
	registry := NewMetricRegistry(sender, SetPrefix("~test"))
	registry.Start()
	defer registry.Stop()

	registry.PointsTracker().IncValid()
	registry.SpansTracker().IncDropped()
	registry.SpanLogsTracker().IncInvalid()

	registry.Flush()

	assert.Equal(t, map[string]float64{
		"~test.events.dropped":     0.0,
		"~test.events.invalid":     0.0,
		"~test.events.valid":       0.0,
		"~test.histograms.dropped": 0.0,
		"~test.histograms.invalid": 0.0,
		"~test.histograms.valid":   0.0,
		"~test.points.dropped":     0.0,
		"~test.points.invalid":     0.0,
		"~test.points.valid":       1.0,
		"~test.span_logs.dropped":  0.0,
		"~test.span_logs.invalid":  1.0,
		"~test.span_logs.valid":    0.0,
		"~test.spans.dropped":      1.0,
		"~test.spans.invalid":      0.0,
		"~test.spans.valid":        0.0,
	}, sender.deltaCounters)
}

type mockSender struct {
	metrics       map[string]float64
	deltaCounters map[string]float64
}

func (m *mockSender) SendMetric(name string, value float64, _ int64, _ string, _ map[string]string) error {
	if m.metrics == nil {
		m.metrics = make(map[string]float64)
	}
	m.metrics[name] = value
	return nil
}

func (m *mockSender) SendDeltaCounter(name string, value float64, _ string, _ map[string]string) error {
	if m.deltaCounters == nil {
		m.deltaCounters = make(map[string]float64)
	}
	m.deltaCounters[name] = value
	return nil
}
