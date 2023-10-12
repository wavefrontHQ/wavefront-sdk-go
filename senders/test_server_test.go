package senders

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

func startTestServer(useTLS bool) *testServer {
	ts := &testServer{}
	handler := http.NewServeMux()
	handler.HandleFunc("/api/v2/event", ts.EventAPIEndpoint)
	handler.HandleFunc("/", ts.ReportEndpoint)
	if useTLS {
		ts.httpServer = httptest.NewTLSServer(handler)
	} else {
		ts.httpServer = httptest.NewServer(handler)
	}
	ts.URL = ts.httpServer.URL
	return ts
}

type testServer struct {
	MetricLines []string
	EventLines  []string
	AuthHeaders []string
	httpServer  *httptest.Server
	URL         string
	RequestURLs []string
}

func (s *testServer) TLSConfig() *tls.Config {
	certPool := x509.NewCertPool()
	certPool.AddCert(s.httpServer.Certificate())
	return &tls.Config{
		RootCAs: certPool,
	}
}

func (s *testServer) ReportEndpoint(writer http.ResponseWriter, request *http.Request) {
	newLines, err := decodeLines(request)
	if err != nil {
		writer.WriteHeader(500)
		return
	}
	s.MetricLines = append(s.MetricLines, newLines...)
	s.AuthHeaders = append(s.AuthHeaders, request.Header.Get("Authorization"))
	s.RequestURLs = append(s.RequestURLs, request.URL.String())
	writer.WriteHeader(200)
}

func (s *testServer) EventAPIEndpoint(writer http.ResponseWriter, request *http.Request) {
	fmt.Println(request.URL.Path)
	newLines, err := decodeLines(request)
	if err != nil {
		writer.WriteHeader(500)
		return
	}
	s.EventLines = append(s.EventLines, newLines...)
	s.AuthHeaders = append(s.AuthHeaders, request.Header.Get("Authorization"))
	s.RequestURLs = append(s.RequestURLs, request.URL.String())
	writer.WriteHeader(200)
}

func decodeLines(request *http.Request) ([]string, error) {
	var metricLines []string
	var bodyReader io.Reader
	defer request.Body.Close()
	if request.Header.Get("Content-Encoding") == "gzip" {
		r, err := gzip.NewReader(request.Body)
		if err != nil {
			return metricLines, err
		}
		defer r.Close()
		bodyReader = r
	} else {
		bodyReader = request.Body
	}

	scanner := bufio.NewScanner(bodyReader)
	for scanner.Scan() {
		line := scanner.Text()
		metricLines = append(metricLines, line)
	}
	if scanner.Err() != nil {
		return metricLines, scanner.Err()
	}
	return metricLines, nil
}

func (s *testServer) Close() {
	s.httpServer.Close()
}

func (s *testServer) hasReceivedLine(lineSubstring string) bool {
	internalMetricFound := false
	for _, line := range s.MetricLines {
		if strings.Contains(line, lineSubstring) {
			internalMetricFound = true
		}
	}
	return internalMetricFound
}
