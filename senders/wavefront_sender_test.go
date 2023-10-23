package senders

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
	"github.com/wavefronthq/wavefront-sdk-go/internal/sdkmetrics"
)

func TestWavefrontSender_SendMetric(t *testing.T) {
	registry := &mockRegistry{}
	pointHandler := &mockHandler{}
	histoHandler := &mockHandler{}
	spanHandler := &mockHandler{}
	spanLogHandler := &mockHandler{}
	eventHandler := &mockHandler{}
	sender := realSender{
		defaultSource:    "test",
		pointHandler:     pointHandler,
		histoHandler:     histoHandler,
		spanHandler:      spanHandler,
		spanLogHandler:   spanLogHandler,
		eventHandler:     eventHandler,
		internalRegistry: registry,
		proxy:            false,
	}

	assert.NoError(t, sender.SendMetric("foo", 20, 0, "test", nil))
	assert.Equal(t, 1, registry.PointsTracker().(*simpleTracker).valid)
	assert.Equal(t, "\"foo\" 20 source=\"test\"\n", pointHandler.Lines[0])
	pointHandler.Reset()

	assert.Error(t, sender.SendMetric("", 21, 0, "test", nil))
	assert.Equal(t, 1, registry.PointsTracker().(*simpleTracker).invalid)
	assert.Empty(t, pointHandler.Lines)

	pointHandler.Error = fmt.Errorf("fake error")

	assert.Error(t, sender.SendMetric("foo", 21, 0, "test", nil))
	assert.Equal(t, 1, registry.PointsTracker().(*simpleTracker).dropped)
	assert.Equal(t, "\"foo\" 21 source=\"test\"\n", pointHandler.Lines[0])
}

func TestWavefrontSender_SendDeltaCounter(t *testing.T) {
	registry := &mockRegistry{}
	pointHandler := &mockHandler{}
	histoHandler := &mockHandler{}
	spanHandler := &mockHandler{}
	spanLogHandler := &mockHandler{}
	eventHandler := &mockHandler{}
	sender := realSender{
		defaultSource:    "test",
		pointHandler:     pointHandler,
		histoHandler:     histoHandler,
		spanHandler:      spanHandler,
		spanLogHandler:   spanLogHandler,
		eventHandler:     eventHandler,
		internalRegistry: registry,
		proxy:            false,
	}

	assert.NoError(t, sender.SendDeltaCounter("foo", 20.0, "test", nil))
	assert.Equal(t, 1, registry.PointsTracker().(*simpleTracker).valid)
	assert.Equal(t, "\"∆foo\" 20 source=\"test\"\n", pointHandler.Lines[0])
	pointHandler.Reset()

	assert.NoError(t, sender.SendDeltaCounter("∆foo", 20.0, "test", nil))
	assert.Equal(t, 2, registry.PointsTracker().(*simpleTracker).valid)
	assert.Equal(t, "\"∆foo\" 20 source=\"test\"\n", pointHandler.Lines[0])
	pointHandler.Reset()

	assert.Error(t, sender.SendDeltaCounter("", 21.0, "test", nil))
	assert.Equal(t, 1, registry.PointsTracker().(*simpleTracker).invalid)
	assert.Empty(t, pointHandler.Lines)

	pointHandler.Error = fmt.Errorf("fake error")

	assert.Error(t, sender.SendDeltaCounter("foo", 21, "test", nil))
	assert.Equal(t, 1, registry.PointsTracker().(*simpleTracker).dropped)
	assert.Equal(t, "\"∆foo\" 21 source=\"test\"\n", pointHandler.Lines[0])
}

func TestWavefrontSender_SendDistribution(t *testing.T) {
	registry := &mockRegistry{}
	pointHandler := &mockHandler{}
	histoHandler := &mockHandler{}
	spanHandler := &mockHandler{}
	spanLogHandler := &mockHandler{}
	eventHandler := &mockHandler{}
	sender := realSender{
		defaultSource:    "test",
		pointHandler:     pointHandler,
		histoHandler:     histoHandler,
		spanHandler:      spanHandler,
		spanLogHandler:   spanLogHandler,
		eventHandler:     eventHandler,
		internalRegistry: registry,
		proxy:            false,
	}

	centroids := []histogram.Centroid{{Value: 0, Count: 0}, {Value: 200, Count: 300}}
	granularities := map[histogram.Granularity]bool{
		8:  false,
		10: true,
	}

	assert.NoError(t, sender.SendDistribution("foo", centroids, granularities, 0, "test", nil))
	assert.Equal(t, 1, registry.HistogramsTracker().(*simpleTracker).valid)
	assert.Regexp(t, "^!D ([#0-9 ]*) \"foo\" source=\"test\"\n$", histoHandler.Lines[0])
	assert.Contains(t, histoHandler.Lines[0], "#0 0")
	assert.Contains(t, histoHandler.Lines[0], "#300 200")
	histoHandler.Reset()

	assert.Error(t, sender.SendDistribution("", centroids, granularities, 0, "test", nil))
	assert.Equal(t, 1, registry.HistogramsTracker().(*simpleTracker).invalid)
	assert.Empty(t, histoHandler.Lines)

	histoHandler.Error = fmt.Errorf("fake error")

	assert.Error(t, sender.SendDistribution("foo", centroids, granularities, 0, "test", nil))
	assert.Equal(t, 1, registry.HistogramsTracker().(*simpleTracker).dropped)
	assert.Regexp(t, "^!D ([#0-9 ]*) \"foo\" source=\"test\"\n$", histoHandler.Lines[0])
	assert.Contains(t, histoHandler.Lines[0], "#0 0")
	assert.Contains(t, histoHandler.Lines[0], "#300 200")
}

func TestWavefrontSender_SendSpan(t *testing.T) {
	registry := &mockRegistry{}
	pointHandler := &mockHandler{}
	histoHandler := &mockHandler{}
	spanHandler := &mockHandler{}
	spanLogHandler := &mockHandler{}
	eventHandler := &mockHandler{}
	sender := realSender{
		defaultSource:    "test",
		pointHandler:     pointHandler,
		histoHandler:     histoHandler,
		spanHandler:      spanHandler,
		spanLogHandler:   spanLogHandler,
		eventHandler:     eventHandler,
		internalRegistry: registry,
		proxy:            false,
	}

	traceID := "28e09666-9610-4690-a908-5298d95551ad"
	spanID := "28b0ad93-58f5-4efe-a68b-7b7a84c8ace8"

	assert.NoError(t, sender.SendSpan(
		"foo",
		200, 2000,
		"test",
		traceID,
		spanID,
		[]string{"pat", "marty"},
		[]string{"gloria"},
		[]SpanTag{},
		[]SpanLog{},
	))
	assert.Equal(t, 1, registry.SpansTracker().(*simpleTracker).valid)
	assert.Equal(t, "\"foo\" source=\"test\" traceId=28e09666-9610-4690-a908-5298d95551ad spanId=28b0ad93-58f5-4efe-a68b-7b7a84c8ace8 parent=pat parent=marty followsFrom=gloria 200 2000\n", spanHandler.Lines[0])
	spanHandler.Reset()

	assert.Error(t, sender.SendSpan(
		"",
		200, 2000,
		"test",
		traceID,
		spanID,
		[]string{"pat", "marty"},
		[]string{"gloria"},
		[]SpanTag{},
		[]SpanLog{},
	))
	assert.Equal(t, 1, registry.SpansTracker().(*simpleTracker).invalid)
	assert.Empty(t, spanHandler.Lines)

	spanHandler.Error = fmt.Errorf("fake error")
	assert.Error(t, sender.SendSpan(
		"foo",
		2000, 400,
		"test",
		traceID,
		spanID,
		[]string{"pat", "marty"},
		[]string{"gloria"},
		[]SpanTag{},
		[]SpanLog{},
	))
	assert.Equal(t, 1, registry.SpansTracker().(*simpleTracker).dropped)
	assert.Equal(t, "\"foo\" source=\"test\" traceId=28e09666-9610-4690-a908-5298d95551ad spanId=28b0ad93-58f5-4efe-a68b-7b7a84c8ace8 parent=pat parent=marty followsFrom=gloria 2000 400\n", spanHandler.Lines[0])
}

func TestWavefrontSender_SendSpan_SpanLogs(t *testing.T) {
	registry := &mockRegistry{}
	pointHandler := &mockHandler{}
	histoHandler := &mockHandler{}
	spanHandler := &mockHandler{}
	spanLogHandler := &mockHandler{}
	eventHandler := &mockHandler{}
	sender := realSender{
		defaultSource:    "test",
		pointHandler:     pointHandler,
		histoHandler:     histoHandler,
		spanHandler:      spanHandler,
		spanLogHandler:   spanLogHandler,
		eventHandler:     eventHandler,
		internalRegistry: registry,
		proxy:            false,
	}

	traceID := "28e09666-9610-4690-a908-5298d95551ad"
	spanID := "28b0ad93-58f5-4efe-a68b-7b7a84c8ace8"

	assert.NoError(t, sender.SendSpan(
		"foo",
		200, 2000,
		"test",
		traceID,
		spanID,
		[]string{"pat", "marty"},
		[]string{"gloria"},
		[]SpanTag{},
		[]SpanLog{
			{Timestamp: 10_000, Fields: map[string]string{"type": "birch"}},
			{Timestamp: 20_000, Fields: map[string]string{"type": "sycamore"}},
		},
	))
	assert.Equal(t, 1, registry.SpanLogsTracker().(*simpleTracker).valid)
	assert.Equal(
		t,
		"{\"traceId\":\"28e09666-9610-4690-a908-5298d95551ad\",\"spanId\":\"28b0ad93-58f5-4efe-a68b-7b7a84c8ace8\",\"logs\":[{\"timestamp\":10000,\"fields\":{\"type\":\"birch\"}},{\"timestamp\":20000,\"fields\":{\"type\":\"sycamore\"}}],\"span\":\"\\\"foo\\\" source=\\\"test\\\" traceId=28e09666-9610-4690-a908-5298d95551ad spanId=28b0ad93-58f5-4efe-a68b-7b7a84c8ace8 parent=pat parent=marty followsFrom=gloria \\\"_spanLogs\\\"=\\\"true\\\" 200 2000\\n\"}\n",
		spanLogHandler.Lines[0],
	)
	spanLogHandler.Reset()

	spanLogHandler.Error = fmt.Errorf("fake error")
	assert.Error(t, sender.SendSpan(
		"foo",
		2000, 400,
		"test",
		traceID,
		spanID,
		[]string{"pat", "marty"},
		[]string{"gloria"},
		[]SpanTag{},
		[]SpanLog{
			{Timestamp: 30_000, Fields: map[string]string{"type": "douglas fir"}},
			{Timestamp: 40_000, Fields: map[string]string{"type": "mountain hemlock"}},
		},
	))
	assert.Equal(t, 1, registry.SpanLogsTracker().(*simpleTracker).dropped)
	assert.Equal(t,
		"{\"traceId\":\"28e09666-9610-4690-a908-5298d95551ad\",\"spanId\":\"28b0ad93-58f5-4efe-a68b-7b7a84c8ace8\",\"logs\":[{\"timestamp\":30000,\"fields\":{\"type\":\"douglas fir\"}},{\"timestamp\":40000,\"fields\":{\"type\":\"mountain hemlock\"}}],\"span\":\"\\\"foo\\\" source=\\\"test\\\" traceId=28e09666-9610-4690-a908-5298d95551ad spanId=28b0ad93-58f5-4efe-a68b-7b7a84c8ace8 parent=pat parent=marty followsFrom=gloria \\\"_spanLogs\\\"=\\\"true\\\" 2000 400\\n\"}\n",
		spanLogHandler.Lines[0],
	)
}

func TestWavefrontSender_SendEventWithProxyFalse(t *testing.T) {
	registry := &mockRegistry{}
	pointHandler := &mockHandler{}
	histoHandler := &mockHandler{}
	spanHandler := &mockHandler{}
	spanLogHandler := &mockHandler{}
	eventHandler := &mockHandler{}
	sender := realSender{
		defaultSource:    "test",
		pointHandler:     pointHandler,
		histoHandler:     histoHandler,
		spanHandler:      spanHandler,
		spanLogHandler:   spanLogHandler,
		eventHandler:     eventHandler,
		internalRegistry: registry,
		proxy:            false,
	}

	assert.NoError(t, sender.SendEvent(
		"foo", 200, 400, "test", nil))
	assert.Equal(t, 1, registry.EventsTracker().(*simpleTracker).valid)
	assert.Equal(t, "{\"annotations\":{},\"endTime\":400000,\"hosts\":[\"test\"],\"name\":\"foo\",\"startTime\":200000}", eventHandler.Lines[0])
	eventHandler.Reset()
	eventHandler.Error = fmt.Errorf("fake error")

	assert.Error(t, sender.SendEvent("foo", 2100, 4100, "test", nil))
	assert.Equal(t, 1, registry.EventsTracker().(*simpleTracker).dropped)
	assert.Equal(t, "{\"annotations\":{},\"endTime\":4100000,\"hosts\":[\"test\"],\"name\":\"foo\",\"startTime\":2100000}", eventHandler.Lines[0])
}

func TestWavefrontSender_SendEventWithProxyTrue(t *testing.T) {
	registry := &mockRegistry{}
	pointHandler := &mockHandler{}
	histoHandler := &mockHandler{}
	spanHandler := &mockHandler{}
	spanLogHandler := &mockHandler{}
	eventHandler := &mockHandler{}
	sender := realSender{
		defaultSource:    "test",
		pointHandler:     pointHandler,
		histoHandler:     histoHandler,
		spanHandler:      spanHandler,
		spanLogHandler:   spanLogHandler,
		eventHandler:     eventHandler,
		internalRegistry: registry,
		proxy:            true,
	}

	assert.NoError(t, sender.SendEvent(
		"foo", 200, 400, "test", nil))
	assert.Equal(t, 1, registry.EventsTracker().(*simpleTracker).valid)
	assert.Equal(t, "@Event 200000 400000 \"foo\" host=\"test\"\n", eventHandler.Lines[0])
	eventHandler.Reset()
	eventHandler.Error = fmt.Errorf("fake error")

	assert.Error(t, sender.SendEvent("foo", 2100, 4100, "test", nil))
	assert.Equal(t, 1, registry.EventsTracker().(*simpleTracker).dropped)
	assert.Equal(t, "@Event 2100000 4100000 \"foo\" host=\"test\"\n", eventHandler.Lines[0])
}

type mockHandler struct {
	Error error
	Lines []string
}

func (m *mockHandler) HandleLine(line string) error {
	m.Lines = append(m.Lines, line)
	return m.Error
}

func (m *mockHandler) Start() {
}

func (m *mockHandler) Stop() {
}

func (m *mockHandler) Flush() error {
	return m.Error
}

func (m *mockHandler) GetFailureCount() int64 {
	return 0
}

func (m *mockHandler) Reset() {
	m.Lines = nil
	m.Error = nil
}

type simpleTracker struct {
	valid   int
	invalid int
	dropped int
}

func (s *simpleTracker) IncValid() {
	s.valid++
}

func (s *simpleTracker) IncInvalid() {
	s.invalid++
}

func (s *simpleTracker) IncDropped() {
	s.dropped++
}

type mockRegistry struct {
	pointsTracker     *simpleTracker
	histogramsTracker *simpleTracker
	spansTracker      *simpleTracker
	spansLogsTracker  *simpleTracker
	eventsTracker     *simpleTracker
}

func (m *mockRegistry) Flush() {
}

func (m *mockRegistry) Start() {
}

func (m *mockRegistry) Stop() {
}

func (m *mockRegistry) PointsTracker() sdkmetrics.SuccessTracker {
	if m.pointsTracker == nil {
		m.pointsTracker = &simpleTracker{}
	}
	return m.pointsTracker
}

func (m *mockRegistry) HistogramsTracker() sdkmetrics.SuccessTracker {
	if m.histogramsTracker == nil {
		m.histogramsTracker = &simpleTracker{}
	}
	return m.histogramsTracker
}

func (m *mockRegistry) SpansTracker() sdkmetrics.SuccessTracker {
	if m.spansTracker == nil {
		m.spansTracker = &simpleTracker{}
	}
	return m.spansTracker
}

func (m *mockRegistry) SpanLogsTracker() sdkmetrics.SuccessTracker {
	if m.spansLogsTracker == nil {
		m.spansLogsTracker = &simpleTracker{}
	}
	return m.spansLogsTracker
}

func (m *mockRegistry) EventsTracker() sdkmetrics.SuccessTracker {
	if m.eventsTracker == nil {
		m.eventsTracker = &simpleTracker{}
	}
	return m.eventsTracker
}

func (m *mockRegistry) NewGauge(string, func() int64) *sdkmetrics.FunctionalGauge {
	return &sdkmetrics.FunctionalGauge{}
}
