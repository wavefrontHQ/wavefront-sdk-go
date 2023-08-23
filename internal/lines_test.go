package internal

import (
	"fmt"
	"github.com/wavefronthq/wavefront-sdk-go/internal/auth"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeReporter struct {
	errorCode int
	error     error
}

func (reporter *fakeReporter) Report(string, string) (*http.Response, error) {
	if reporter.error != nil {
		return nil, reporter.error
	}
	if reporter.errorCode != 0 {
		return &http.Response{StatusCode: reporter.errorCode}, nil
	}
	return &http.Response{StatusCode: 200}, nil
}

func (reporter *fakeReporter) ReportEvent(string) (*http.Response, error) {
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

func TestHandleLine_OnAuthError_DoNotBuffer(t *testing.T) {
	lh := makeLineHandler(100, 10) // cap: 100, batchSize: 10
	lh.Reporter = &fakeReporter{
		error: auth.NewAuthError(fmt.Errorf("fake auth error that shouldn't be buffered")),
	}
	assert.NoError(t, lh.HandleLine("this is a metric, but CSP is down, or my credentials are wrong"))
	assert.Error(t, lh.Flush())
	checkLength(lh.buffer, 0, "", t)
	lh.Reporter = &fakeReporter{
		error: fmt.Errorf("error that should be buffered"),
	}
	assert.NoError(t, lh.HandleLine("this is a metric, but it was a network timeout or something like that"))
	assert.Error(t, lh.Flush())
	checkLength(lh.buffer, 1, "", t)
}

func TestFlush(t *testing.T) {
	lh := makeLineHandler(100, 10) // cap: 100, batchSize: 10

	addLines(lh, 100, 100, t)
	assert.NoError(t, lh.Flush())
	assert.Equal(t, 90, len(lh.buffer), "error flushing lines")

	e := fmt.Errorf("error reporting points")
	lh.Reporter = &fakeReporter{error: e}
	assert.Error(t, lh.Flush())
	assert.Equal(t, 90, len(lh.buffer), "error flushing lines")

	lh.Reporter = &fakeReporter{}
	lh.buffer = make(chan string, 100)
	addLines(lh, 5, 5, t)
	assert.NoError(t, lh.Flush())
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
