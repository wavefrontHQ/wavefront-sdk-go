// Package internal offers helper interfaces that are internal to the Wavefront Go SDK.
// Interfaces within this package are not guaranteed to be backwards compatible between releases.
package internal

import "net/http"

// Reporter is an interface for reporting data to a Wavefront service.
type Reporter interface {
	Report(format string, pointLines string) (*http.Response, error)
}

type Flusher interface {
	Flush() error
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
	FlushWithThrottling() error
	GetFailureCount() int64
	Format() string
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
