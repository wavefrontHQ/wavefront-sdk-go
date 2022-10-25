package senders

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
