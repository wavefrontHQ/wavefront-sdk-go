package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

type testServer struct {
	MetricLines    []string
	httpServer     *httptest.Server
	URL            string
	LastRequestURL string
}

func (s *testServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	newLines, err := decodeMetricLines(request)
	if err != nil {
		writer.WriteHeader(500)
	}
	s.MetricLines = append(s.MetricLines, newLines...)
	s.LastRequestURL = request.URL.String()
	time.Sleep(9 * time.Second)
	writer.WriteHeader(200)
	log.Printf("%v", newLines)
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

func main() {
	certFilePath := os.Args[1]
	keyFilePath := os.Args[2]
	fmt.Printf(`
	Starting Mock Wavefront Server on https://localhost:9999
	certFilePath: %s
	keyFilePath: %s
`, certFilePath, keyFilePath)
	server := testServer{}
	http.Handle("/report", &server)

	log.Panicln(http.ListenAndServeTLS(
		"localhost:9999",
		certFilePath,
		keyFilePath,
		nil,
	))
}
