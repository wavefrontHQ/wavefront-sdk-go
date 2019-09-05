package internal

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var (
	client    = &http.Client{Timeout: time.Second * 10}
	errReport = errors.New("error: invalid Format or points")
)

const (
	contentType     = "Content-Type"
	contentEncoding = "Content-Encoding"
	authzHeader     = "Authorization"
	bearer          = "Bearer "
	gzipFormat      = "gzip"
	formatKey       = "f"
)

// The implementation of a Reporter that reports points directly to a Wavefront server.
type directReporter struct {
	serverURL   string
	token       string
	endpoint    string
	contentType string
	gzip        bool
}

// ReporterOption allows Reporter configuration
type ReporterOption func(*directReporter)

// SetEndpoint set the Reporter endpoint, '/report' by default
func SetEndpoint(endpoint string) ReporterOption {
	return func(reporter *directReporter) {
		reporter.endpoint = endpoint
	}
}

// SetContentType set the Reporter contentType, 'application/octet-stream' by default
func SetContentType(contentType string) ReporterOption {
	return func(reporter *directReporter) {
		reporter.contentType = contentType
	}
}

// EnableGZip set the Reporter gzip compression, 'true' by default
func EnableGZip(gzip bool) ReporterOption {
	return func(reporter *directReporter) {
		reporter.gzip = gzip
	}
}

// NewDirectReporter create a Reporter
func NewDirectReporter(server string, token string, setters ...ReporterOption) Reporter {
	dr := &directReporter{
		serverURL:   server,
		token:       token,
		endpoint:    "/report",
		contentType: "application/octet-stream",
		gzip:        true,
	}
	for _, setter := range setters {
		setter(dr)
	}
	return dr
}

func (reporter directReporter) Report(format string, pointLines string) (*http.Response, error) {
	if format == "" || pointLines == "" {
		return nil, errReport
	}

	var buf io.Reader
	if reporter.gzip {
		var gzbuf bytes.Buffer
		zw := gzip.NewWriter(&gzbuf)
		_, err := zw.Write([]byte(pointLines))
		if err != nil {
			zw.Close()
			return nil, err
		}
		if err = zw.Close(); err != nil {
			return nil, err
		}
		buf = &gzbuf
	} else {
		buf = strings.NewReader(pointLines)
	}

	apiURL := reporter.serverURL + reporter.endpoint
	req, err := http.NewRequest("POST", apiURL, buf)
	if err != nil {
		return &http.Response{}, err
	}

	req.Header.Set(contentType, reporter.contentType)
	req.Header.Set(authzHeader, bearer+reporter.token)
	if reporter.gzip {
		req.Header.Set(contentEncoding, gzipFormat)
	}

	if format != "event" {
		q := req.URL.Query()
		q.Add(formatKey, format)
		req.URL.RawQuery = q.Encode()
	}

	resp, err := client.Do(req)
	if err != nil {
		return resp, err
	}
	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()
	return resp, nil
}

func (reporter directReporter) Server() string {
	return reporter.serverURL
}
