package internal

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-sdk-go/internal/auth"
)

type fakeReporter struct {
	httpResponseStatus int64
	reportCallCount    int64
	error              error
	lines              []string
}

func (reporter *fakeReporter) Report(_ string, lines string) (*http.Response, error) {
	atomic.AddInt64(&reporter.reportCallCount, 1)
	if reporter.error != nil {
		return nil, reporter.error
	}
	status := atomic.LoadInt64(&reporter.httpResponseStatus)
	if status != 0 {
		return &http.Response{StatusCode: int(status)}, nil
	}
	reporter.lines = append(reporter.lines, lines)
	return &http.Response{StatusCode: 200}, nil
}

func (reporter *fakeReporter) ReportCallCount() int {
	return int(atomic.LoadInt64(&reporter.reportCallCount))
}

func (reporter *fakeReporter) ResetReportCallCount() {
	atomic.StoreInt64(&reporter.reportCallCount, int64(0))
}

func (reporter *fakeReporter) SetHTTPStatus(status int) {
	atomic.StoreInt64(&reporter.httpResponseStatus, int64(status))
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

func TestFlushWithThrottling_WhenThrottling_DelayUntilThrottleInterval(t *testing.T) {
	lh := &RealBatchBuilder{
		Reporter:               &fakeReporter{},
		MaxBufferSize:          100,
		BatchSize:              10,
		buffer:                 make(chan string, 100),
		throttleOnBackpressure: true,
		throttledSleepDuration: 1 * time.Second,
	}

	addLines(lh, 100, 100, t)
	lh.Reporter.(*fakeReporter).SetHTTPStatus(406)
	startTime := time.Now().Add(1 * time.Second)
	deadline := startTime.Add(1 * time.Second)
	assert.Error(t, lh.Flush())
	assert.Equal(t, 100, len(lh.buffer))
	assert.WithinRange(t, lh.resumeAt, startTime, deadline)
	lh.Reporter.(*fakeReporter).SetHTTPStatus(0)
	assert.NoError(t, lh.FlushWithThrottling())
	assert.Greater(t, time.Now(), lh.resumeAt)
	assert.Equal(t, 90, len(lh.buffer))
}

func TestBackgroundFlushWithThrottling_WhenThrottling_DelayUntilThrottleInterval(t *testing.T) {
	lh := &RealBatchBuilder{
		Reporter:               &fakeReporter{},
		MaxBufferSize:          100,
		BatchSize:              10,
		buffer:                 make(chan string, 100),
		throttleOnBackpressure: true,
		throttledSleepDuration: 1 * time.Second,
	}

	addLines(lh, 100, 100, t)
	lh.Reporter.(*fakeReporter).SetHTTPStatus(406)
	startTime := time.Now().Add(1 * time.Second)
	deadline := startTime.Add(1 * time.Second)
	assert.Error(t, lh.Flush())
	assert.Equal(t, 100, len(lh.buffer))
	assert.WithinRange(t, lh.resumeAt, startTime, deadline)
	lh.Reporter = &fakeReporter{}
	assert.NoError(t, lh.FlushWithThrottling())
	assert.Greater(t, time.Now(), lh.resumeAt)
	assert.Equal(t, 90, len(lh.buffer))
}

func TestFlushTicker_WhenThrottlingEnabled_AndReceives406Error_ThrottlesRequestsUntilNextSleepDuration(t *testing.T) {
	throttledSleepDuration := 250 * time.Millisecond
	briskTickTime := 50 * time.Millisecond
	lh := &RealBatchBuilder{
		Reporter:               &fakeReporter{},
		MaxBufferSize:          100,
		BatchSize:              10,
		buffer:                 make(chan string, 100),
		throttleOnBackpressure: true,
		throttledSleepDuration: throttledSleepDuration,
	}

	lh.flusher = NewBackgroundFlusher(briskTickTime, lh)
	lh.Start()
	addLines(lh, 100, 100, t)

	twoTicksOfTheTicker := 2 * briskTickTime
	time.Sleep(twoTicksOfTheTicker + 10*time.Millisecond)

	assert.Equal(t, 2, lh.Reporter.(*fakeReporter).ReportCallCount())
	lh.Reporter.(*fakeReporter).ResetReportCallCount()

	lh.Reporter.(*fakeReporter).SetHTTPStatus(406)
	time.Sleep(twoTicksOfTheTicker)

	assert.Equal(t, 1, lh.Reporter.(*fakeReporter).ReportCallCount())

	lh.Stop()
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

func addLines(lh *RealBatchBuilder, linesToAdd int, expectedLen int, t *testing.T) {
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

func makeLineHandler(bufSize, batchSize int) *RealBatchBuilder {
	return &RealBatchBuilder{
		Reporter:      &fakeReporter{},
		MaxBufferSize: bufSize,
		BatchSize:     batchSize,
		buffer:        make(chan string, bufSize),
	}
}
