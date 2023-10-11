package senders

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

// Plan:
// Send many metrics via Telegraf's Wavefront Output Plugin (through wavefront-sdk-go) to a destination server.
// In the TAS Integration tile case, this server would be the Wavefront Proxy.
// In this attempt to reproduce, this server would be sleepyTestServer.

// Steps:
// 1. brew install telegraf
// 2. Configure telegraf.conf [[outputs.wavefront]] to point to the sleepyTestServer.
// Note: This may require knowing the the port on which sleepyTestServer is listening.
// To get this port, it may be worth creating sleepyTestServer using an `http` package (rather than an `httptest`) package method.
// 3. Write code to send many metrics to the test server.
// 4. Tail the Telegraf logs to see if the error appears.

func startSleepyTestServer(useTLS bool, secondsToSleep int) *sleepyTestServer {
	handler := &sleepyTestServer{}

	if useTLS {
		handler.httpServer = httptest.NewTLSServer(handler)
	} else {
		handler.httpServer = httptest.NewServer(handler)
	}
	handler.URL = handler.httpServer.URL
	return handler
}

type sleepyTestServer struct {
	MetricLines    []string
	AuthHeaders    []string
	httpServer     *httptest.Server
	URL            string
	LastRequestURL string
}

func (s *sleepyTestServer) TLSConfig() *tls.Config {
	certpool := x509.NewCertPool()
	certpool.AddCert(s.httpServer.Certificate())
	return &tls.Config{
		RootCAs: certpool,
	}
}

func (s *sleepyTestServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	time.Sleep(5 * time.Second)
	newLines, err := decodeMetricLines(request)
	if err != nil {
		writer.WriteHeader(500)
	}
	s.MetricLines = append(s.MetricLines, newLines...)
	s.AuthHeaders = append(s.AuthHeaders, request.Header.Get("Authorization"))
	s.LastRequestURL = request.URL.String()
	writer.WriteHeader(200)
}

func (s *sleepyTestServer) Close() {
	s.httpServer.Close()
}

func (s *sleepyTestServer) hasReceivedLine(lineSubstring string) bool {
	internalMetricFound := false
	for _, line := range s.MetricLines {
		if strings.Contains(line, lineSubstring) {
			internalMetricFound = true
		}
	}
	return internalMetricFound
}

func main() {
	testServer := startSleepyTestServer(false, 60)
	defer testServer.Close()
	sender, err := NewSender(testServer.URL)

	if err != nil {
		// handle error
	}
	// TODO: Send many metrics, not just one.
	sender.SendMetric("my metric", 20, 0, "localhost", nil)
	sender.Flush()
}
