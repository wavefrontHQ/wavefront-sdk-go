// Package internal offers helper interfaces that are internal to the Wavefront Go SDK.
// Interfaces within this package are not guaranteed to be backwards compatible between releases.
package internal

import "net/http"

// Reporter is an interface for reporting data to a Wavefront service.
type Reporter interface {
	Report(format string, pointLines string) (*http.Response, error)
}

type Flusher interface {
	// Deprecated: the Flush method in this SDK behaves differently from typical methods of the same name.
	// It flushes a single batch or lines, rather than all lines waiting to be sent.
	// For a full flush, use FlushAll. For the legacy Flush behavior, use FlushOneBatch.
	// Flush is now an alias for FlushOneBatch.
	Flush() error
	FlushAll() error
	FlushOneBatch() error
	GetFailureCount() int64
	Start()
}

type ConnectionHandler interface {
	Connect() error
	Connected() bool
	Close()
	SendData(lines string) error

	Flusher
}

type LineHandler interface {
	HandleLine(line string) error
	Start()
	Stop()
	Flush() error
	GetFailureCount() int64
}

const (
	contentType     = "Content-Type"
	contentEncoding = "Content-Encoding"
	gzipFormat      = "gzip"

	octetStream     = "application/octet-stream"
	applicationJSON = "application/json"

	reportEndpoint = "/report"
	eventEndpoint  = "/api/v2/event"

	formatKey = "f"
)

const formatError stringError = "error: invalid Format or points"

type stringError string

func (e stringError) Error() string {
	return string(e)
}
