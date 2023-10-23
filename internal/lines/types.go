package lines

import "net/http"

// Reporter is an interface for reporting data to a Wavefront service.
type Reporter interface {
	Report(format string, pointLines string) (*http.Response, error)
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

type Flusher interface {
	Start()
	Stop()
}