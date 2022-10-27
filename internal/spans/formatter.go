package spans

import (
	"errors"
	"fmt"
	"github.com/wavefronthq/wavefront-sdk-go/internal"
	"strconv"
)

// Gets a span line in the Wavefront span data format:
// <tracingSpanName> source=<source> [pointTags] <start_millis> <duration_milli_seconds>
// Example:
// "getAllUsers source=localhost traceId=7b3bf470-9456-11e8-9eb6-529269fb1459 spanId=0313bafe-9457-11e8-9eb6-529269fb1459
//
//	parent=2f64e538-9457-11e8-9eb6-529269fb1459 application=Wavefront http.method=GET 1533531013 343500"
func SpanLine(name string, startMillis, durationMillis int64, source, traceId, spanId string, parents, followsFrom []string, tags []SpanTag, spanLogs []SpanLog, defaultSource string) (string, error) {
	if name == "" {
		return "", errors.New("span name cannot be empty")
	}

	if source == "" {
		source = defaultSource
	}

	if !internal.IsUUIDFormat(traceId) {
		return "", fmt.Errorf("traceId is not in UUID format: span=%s traceId=%s", name, traceId)
	}
	if !internal.IsUUIDFormat(spanId) {
		return "", fmt.Errorf("spanId is not in UUID format: span=%s spanId=%s", name, spanId)
	}

	sb := internal.GetBuffer()
	defer internal.PutBuffer(sb)

	sb.WriteString(internal.SanitizeValue(name))
	sb.WriteString(" source=")
	sb.WriteString(internal.SanitizeValue(source))
	sb.WriteString(" traceId=")
	sb.WriteString(traceId)
	sb.WriteString(" spanId=")
	sb.WriteString(spanId)

	for _, parent := range parents {
		sb.WriteString(" parent=")
		sb.WriteString(parent)
	}

	for _, item := range followsFrom {
		sb.WriteString(" followsFrom=")
		sb.WriteString(item)
	}

	if len(spanLogs) > 0 {
		sb.WriteString(" ")
		sb.WriteString(strconv.Quote(internal.Sanitize("_spanLogs")))
		sb.WriteString("=")
		sb.WriteString(strconv.Quote(internal.Sanitize("true")))
	}

	for _, tag := range tags {
		if tag.Key == "" {
			return "", fmt.Errorf("tag keys cannot be empty: span=%s", name)
		}
		if tag.Value == "" {
			return "", fmt.Errorf("tag values cannot be empty: span=%s tag=%s", name, tag.Key)
		}
		sb.WriteString(" ")
		sb.WriteString(strconv.Quote(internal.Sanitize(tag.Key)))
		sb.WriteString("=")
		sb.WriteString(internal.SanitizeValue(tag.Value))
	}
	sb.WriteString(" ")
	sb.WriteString(strconv.FormatInt(startMillis, 10))
	sb.WriteString(" ")
	sb.WriteString(strconv.FormatInt(durationMillis, 10))
	sb.WriteString("\n")

	return sb.String(), nil
}
