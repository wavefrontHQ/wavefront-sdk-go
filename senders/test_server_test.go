package senders

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
)

func startTestServer() *testServer {
	handler := &testServer{}
	server := httptest.NewServer(handler)
	handler.httpServer = server
	handler.URL = server.URL
	return handler

}

func startTLSTestServer() *testServer {
	handler := &testServer{}
	server := httptest.NewTLSServer(handler)
	handler.httpServer = server
	handler.URL = server.URL
	return handler
}

type testServer struct {
	MetricLines []string
	httpServer  *httptest.Server
	URL         string
}

func (s *testServer) TLSConfig() tls.Config {
	certpool := x509.NewCertPool()
	certpool.AddCert(s.httpServer.Certificate())
	return tls.Config{
		RootCAs: certpool,
	}
}

func (s *testServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	newLines, err := decodeMetricLines(request)
	if err != nil {
		writer.WriteHeader(500)
	}
	s.MetricLines = append(s.MetricLines, newLines...)
	writer.WriteHeader(200)
}

func decodeMetricLines(request *http.Request) ([]string, error) {
	var metricLines []string
	reader, err := gzip.NewReader(request.Body)
	if err != nil {
		return metricLines, err
	}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		metricLines = append(metricLines, line)
	}
	if scanner.Err() != nil {
		reader.Close()
		return metricLines, scanner.Err()
	}
	return metricLines, reader.Close()
}

func (s *testServer) Close() {
	s.httpServer.Close()
}
