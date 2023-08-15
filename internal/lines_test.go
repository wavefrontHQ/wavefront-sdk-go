package internal

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeReporter struct {
	raiseError bool
	errorCode  int
}

func (reporter *fakeReporter) Report(format string, pointLines string) (*http.Response, error) {
	if reporter.raiseError {
		return nil, fmt.Errorf("error reporting points")
	}
	if reporter.errorCode != 0 {
		return &http.Response{StatusCode: reporter.errorCode}, nil
	}
	return &http.Response{StatusCode: 200}, nil
}

func (reporter *fakeReporter) ReportEvent(event string) (*http.Response, error) {
	return &http.Response{StatusCode: 200}, nil
}

func TestCapacity(t *testing.T) {
	lh := makeLineHandler(100, 10) // cap: 100, batchSize: 10
	checkLength(lh.buffer, 0, "non-empty lines length", t)

	addLines(lh, 100, 100, t)
	err := lh.HandleLine("dummyLine")
	if err == nil {
		t.Errorf("buffer capacity exceeded but no error")
	}
}

func TestBufferLines(t *testing.T) {
	lh := makeLineHandler(100, 10) // cap: 100, batchSize: 10
	checkLength(lh.buffer, 0, "non-empty lines length", t)

	addLines(lh, 90, 90, t)
	buf := makeBuffer(50)
	lh.bufferLines(buf)
	checkLength(lh.buffer, 100, "error buffering lines", t)

	// clear lines
	lh.buffer = make(chan string, 100)
	checkLength(lh.buffer, 0, "error clearing lines", t)

	addLines(lh, 90, 90, t)
	buf = makeBuffer(5)
	lh.bufferLines(buf)
	checkLength(lh.buffer, 95, "error buffering lines", t)
}

func TestFlush(t *testing.T) {
	lh := makeLineHandler(100, 10) // cap: 100, batchSize: 10

	addLines(lh, 100, 100, t)
	lh.Flush()
	assert.Equal(t, 90, len(lh.buffer), "error flushing lines")

	lh.Reporter = &fakeReporter{raiseError: true}
	lh.Flush()
	assert.Equal(t, 90, len(lh.buffer), "error flushing lines")

	lh.Reporter = &fakeReporter{}
	lh.buffer = make(chan string, 100)
	addLines(lh, 5, 5, t)
	lh.Flush()
	assert.Equal(t, 0, len(lh.buffer), "error flushing lines")
}

func checkLength(buffer chan string, length int, msg string, t *testing.T) {
	if len(buffer) != length {
		t.Errorf("%s. expected: %d actual: %d", msg, length, len(buffer))
	}
}

func addLines(lh *RealLineHandler, linesToAdd int, expectedLen int, t *testing.T) {
	for i := 0; i < linesToAdd; i++ {
		err := lh.HandleLine("dummyLine")
		if err != nil {
			t.Error(err)
		}
	}
	if len(lh.buffer) != expectedLen {
		t.Errorf("error adding lines. expected: %d actual: %d", expectedLen, len(lh.buffer))
	}
}

func makeBuffer(num int) []string {
	buf := make([]string, num)
	for i := 0; i < num; i++ {
		buf[i] = "dummyLine"
	}
	return buf
}

func makeLineHandler(bufSize, batchSize int) *RealLineHandler {
	return &RealLineHandler{
		Reporter:      &fakeReporter{},
		MaxBufferSize: bufSize,
		BatchSize:     batchSize,
		buffer:        make(chan string, bufSize),
	}
}
