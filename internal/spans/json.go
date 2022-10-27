package spans

import "encoding/json"

func SpanLogJSON(traceId, spanId string, spanLogs []SpanLog, span string) (string, error) {
	l := SpanLogs{
		TraceId: traceId,
		SpanId:  spanId,
		Logs:    spanLogs,
		Span:    span,
	}
	out, err := json.Marshal(l)
	if err != nil {
		return "", err
	}
	return string(out[:]) + "\n", nil
}
