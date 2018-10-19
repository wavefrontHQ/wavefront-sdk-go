package internal

import (
	"fmt"
	"net/http"
	"testing"
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

func (reporter *fakeReporter) Server() string {
	return "fake"
}

func TestCapacity(t *testing.T) {
	lh := makeLineHandler(100, 10) // cap: 100, batchSize: 10
	checkLength(lh.lines, 0, "non-empty lines length", t)

	addLines(lh, 100, 100, t)
	err := lh.HandleLine("dummyLine")
	if err == nil {
		t.Errorf("buffer capacity exceeded but no error")
	}
}

func TestBufferLines(t *testing.T) {
	lh := makeLineHandler(100, 10) // cap: 100, batchSize: 10
	checkLength(lh.lines, 0, "non-empty lines length", t)

	addLines(lh, 90, 90, t)
	buf := makeBuffer(50)
	lh.bufferLines(buf)
	checkLength(lh.lines, 100, "error buffering lines", t)

	// clear lines
	lh.lines = nil
	checkLength(lh.lines, 0, "error clearing lines", t)

	addLines(lh, 90, 90, t)
	buf = makeBuffer(5)
	lh.bufferLines(buf)
	checkLength(lh.lines, 95, "error buffering lines", t)
}

func TestFlush(t *testing.T) {
	lh := makeLineHandler(100, 10) // cap: 100, batchSize: 10

	addLines(lh, 100, 100, t)
	lh.Flush()
	checkLength(lh.lines, 90, "error flushing lines", t)

	lh.Reporter = &fakeReporter{raiseError: true}
	lh.Flush()
	checkLength(lh.lines, 90, "error flushing lines", t)
}

func TestBatching(t *testing.T) {
	lh := makeLineHandler(10, 5) // cap: 10, batchSize: 5
	checkLength(lh.linesBatch(), 0, "non-empty batch length", t)

	addLines(lh, 2, 2, t)
	checkLength(lh.linesBatch(), 2, "invalid batch length", t)
	checkLength(lh.lines, 0, "lines not cleared after batching", t)

	addLines(lh, 8, 8, t)
	checkLength(lh.linesBatch(), 5, "invalid batch length", t)
	checkLength(lh.lines, 3, "lines not cleared after batching", t)

	addLines(lh, 2, 5, t)
	checkLength(lh.linesBatch(), 5, "invalid batch length", t)
	checkLength(lh.lines, 0, "lines not cleared after batching", t)
}

func checkLength(buffer []string, length int, msg string, t *testing.T) {
	if len(buffer) != length {
		t.Errorf("%s. expected: %d actual: %d", msg, length, len(buffer))
	}
}

func addLines(lh *LineHandler, linesToAdd int, expectedLen int, t *testing.T) {
	for i := 0; i < linesToAdd; i++ {
		err := lh.HandleLine("dummyLine")
		if err != nil {
			t.Error(err)
		}
	}
	if len(lh.lines) != expectedLen {
		t.Errorf("error adding lines. expected: %d actual: %d", expectedLen, len(lh.lines))
	}
}

func makeBuffer(num int) []string {
	buf := make([]string, num)
	for i := 0; i < num; i++ {
		buf[i] = "dummyLine"
	}
	return buf
}

func makeLineHandler(bufSize, batchSize int) *LineHandler {
	return &LineHandler{
		Reporter:      &fakeReporter{},
		MaxBufferSize: bufSize,
		BatchSize:     batchSize,
	}
}
