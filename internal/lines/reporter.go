package lines

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/wavefronthq/wavefront-sdk-go/internal/auth"
)

const (
	MetricFormat    = "wavefront"
	HistogramFormat = "histogram"
	TraceFormat     = "trace"
	SpanLogsFormat  = "spanLogs"
	EventFormat     = "event"
)

type ReporterOptions struct {
	// max batch of data sent per flush interval. defaults to 10,000. recommended not to exceed 40,000.
	BatchSize int
	// size of internal buffers beyond which received data is dropped.
	// helps with handling brief increases in data and buffering on errors.
	// separate buffers are maintained per data type (metrics, spans and distributions)
	// buffers are not pre-allocated to max size and vary based on actual usage.
	// defaults to 500,000. higher values could use more memory.
	MaxBufferSize int

	// interval (in seconds) at which to flush data to Wavefront. defaults to 1 Second.
	// together with batch size controls the max theoretical throughput of the sender.
	FlushInterval time.Duration
}

// The implementation of a Reporter that reports points directly to a Wavefront server.
type reporter struct {
	serverURL    string
	tokenService auth.Service
	client       *http.Client
}

// NewReporter creates a metrics Reporter
func NewReporter(server string, tokenService auth.Service, client *http.Client) Reporter {
	return &reporter{
		serverURL:    server,
		tokenService: tokenService,
		client:       client,
	}
}

// Report creates and sends a POST to the reportEndpoint with the given pointLines
// nit: format is string as enum/polymorphism
func (reporter reporter) Report(format string, pointLines string) (*http.Response, error) {
	if format == "" || pointLines == "" {
		return nil, formatError
	}

	if format == EventFormat {
		return reporter.reportEvent(pointLines)
	}

	requestBody, err := linesToGzippedBytes(pointLines)
	if err != nil {
		return nil, err
	}

	req, err := reporter.buildRequest(format, requestBody)
	if err != nil {
		return nil, err
	}

	return reporter.execute(req)
}

func linesToGzippedBytes(pointLines string) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write([]byte(pointLines))
	if err != nil {
		_ = zw.Close()
		return nil, err
	}
	if err = zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

func (reporter reporter) buildRequest(format string, body []byte) (*http.Request, error) {
	apiURL := reporter.serverURL + reportEndpoint
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set(contentType, octetStream)
	req.Header.Set(contentEncoding, gzipFormat)

	err = reporter.tokenService.Authorize(req)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add(formatKey, format)
	req.URL.RawQuery = q.Encode()
	return req, nil
}

func (reporter reporter) reportEvent(event string) (*http.Response, error) {
	if event == "" {
		return nil, formatError
	}

	apiURL := reporter.serverURL + eventEndpoint
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(event))
	if err != nil {
		return nil, err
	}

	req.Header.Set(contentType, applicationJSON)

	if reporter.IsDirect() {
		req.Header.Set(contentEncoding, gzipFormat)
	}

	err = reporter.tokenService.Authorize(req)
	if err != nil {
		return nil, err
	}

	return reporter.execute(req)
}

func (reporter reporter) execute(req *http.Request) (*http.Response, error) {
	resp, err := reporter.client.Do(req)
	if err != nil {
		return nil, err
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	defer resp.Body.Close()
	return resp, nil
}

func (reporter reporter) Close() {
	reporter.tokenService.Close()
}

func (reporter reporter) IsDirect() bool {
	return reporter.tokenService.IsDirect()
}
