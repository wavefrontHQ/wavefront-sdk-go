package main

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"time"
)

// Plan:
// Send many metrics via Telegraf's Wavefront Output Plugin (through wavefront-sdk-go) to a destination server.
// In the TAS Integration tile case, this server would be the Wavefront Proxy.
// In this attempt to reproduce, this server would be sleepyTestServer.

// Steps:
// 1. brew install telegraf
// 2. Configure telegraf.conf [[outputs.wavefront]] to point to the sleepyTestServer.
// Note: This may require knowing the port on which sleepyTestServer is listening.
// To get this port, it may be worth creating sleepyTestServer using an `http` package (rather than an `httptest`) package method.
// 3. Write code to send many metrics to the test server.
// 4. Tail the Telegraf logs to see if the error appears.

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
	log.Printf("========running func ServHTTP with url path: %s========\n", request.URL.Path[1:])
	time.Sleep(11 * time.Second)
	newLines, err := decodeMetricLines(request)
	if err != nil {
		writer.WriteHeader(500)
		return
	}
	s.MetricLines = append(s.MetricLines, newLines...)
	s.AuthHeaders = append(s.AuthHeaders, request.Header.Get("Authorization"))
	s.LastRequestURL = request.URL.String()
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

type ConnectionWatcher struct {
	n int64
}

// OnStateChange records open connections in response to connection
// state changes. Set net/http Server.ConnState to this method
// as value.
func (cw *ConnectionWatcher) OnStateChange(conn net.Conn, state http.ConnState) {
	switch state {
	case http.StateNew:
		cw.Add(1)
	case http.StateHijacked, http.StateClosed:
		cw.Add(-1)
	}
	fmt.Printf("Connection state: %s\n", state)
	fmt.Printf("Connection count: %d\n", cw.Count())
}

// Count returns the number of connections at the time
// the call.
func (cw *ConnectionWatcher) Count() int {
	return int(atomic.LoadInt64(&cw.n))
}

// Add adds c to the number of active connections.
func (cw *ConnectionWatcher) Add(c int64) {
	atomic.AddInt64(&cw.n, c)
}

func main() {
	testServer := &sleepyTestServer{}
	cw := &ConnectionWatcher{}
	s := &http.Server{
		Addr:      "localhost:8080",
		Handler:   testServer,
		ConnState: cw.OnStateChange,
	}

	log.Fatal(s.ListenAndServe())
}
