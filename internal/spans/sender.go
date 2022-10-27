package spans

type SpanTag struct {
	Key   string
	Value string
}

type SpanLog struct {
	Timestamp int64             `json:"timestamp"`
	Fields    map[string]string `json:"fields"`
}

// SpanLogs is for internal use only.
type SpanLogs struct {
	TraceId string    `json:"traceId"`
	SpanId  string    `json:"spanId"`
	Logs    []SpanLog `json:"logs"`
	Span    string    `json:"span"`
}

// SpanSender Interface for sending tracing spans to Wavefront
type SpanSender interface {
	// Sends a tracing span to Wavefront.
	// traceId, spanId, parentIds and preceding spanIds are expected to be UUID strings.
	// parents and preceding spans can be empty for a root span.
	// span tag keys can be repeated (example: "user"="foo" and "user"="bar")
	// span logs are currently omitted
	SendSpan(name string, startMillis, durationMillis int64, source, traceId, spanId string, parents, followsFrom []string, tags []SpanTag, spanLogs []SpanLog) error
}
