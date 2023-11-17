package internal

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wavefronthq/wavefront-sdk-go/internal/auth"
	"github.com/wavefronthq/wavefront-sdk-go/internal/sdkmetrics"
)

const (
	metricFormat                  = "wavefront"
	histogramFormat               = "histogram"
	traceFormat                   = "trace"
	spanLogsFormat                = "spanLogs"
	eventFormat                   = "event"
	defaultThrottledSleepDuration = time.Second * 30
)

type RealBatchBuilder struct {
	// keep these two fields as first element of struct
	// to guarantee 64-bit alignment on 32-bit machines.
	// atomic.* functions crash if operands are not 64-bit aligned.
	// See https://github.com/golang/go/issues/599
	failures  int64
	throttled int64

	Reporter      Reporter
	BatchSize     int
	MaxBufferSize int
	format        string

	internalRegistry       sdkmetrics.Registry
	prefix                 string
	throttleOnBackpressure bool
	throttledSleepDuration time.Duration
	mtx                    sync.Mutex

	buffer   chan string
	flusher  BackgroundFlusher
	resumeAt time.Time
}

func (lh *RealLineHandler) Format() string {
	return lh.format
}

var errThrottled = errors.New("error: throttled event creation")

type BatchAccumulatorOption func(*RealBatchBuilder)

func SetRegistry(registry sdkmetrics.Registry) BatchAccumulatorOption {
	return func(handler *RealBatchBuilder) {
		handler.internalRegistry = registry
	}
}

func SetHandlerPrefix(prefix string) BatchAccumulatorOption {
	return func(handler *RealBatchBuilder) {
		handler.prefix = prefix
	}
}

func ThrottleRequestsOnBackpressure() BatchAccumulatorOption {
	return func(handler *RealBatchBuilder) {
		handler.throttleOnBackpressure = true
	}
}

func NewLineHandler(reporter Reporter, format string, flushInterval time.Duration, batchSize, maxBufferSize int, setters ...BatchAccumulatorOption) *RealBatchBuilder {
	lh := &RealBatchBuilder{
		Reporter:               reporter,
		BatchSize:              batchSize,
		MaxBufferSize:          maxBufferSize,
		format:                 format,
		throttledSleepDuration: defaultThrottledSleepDuration,
	}

	lh.buffer = make(chan string, lh.MaxBufferSize)
	lh.flusher = NewBackgroundFlusher(flushInterval, lh)

	for _, setter := range setters {
		setter(lh)
	}

	if lh.internalRegistry != nil {
		lh.internalRegistry.NewGauge(lh.prefix+".queue.size", func() int64 {
			return int64(len(lh.buffer))
		})
		lh.internalRegistry.NewGauge(lh.prefix+".queue.remaining_capacity", func() int64 {
			return int64(lh.MaxBufferSize - len(lh.buffer))
		})
	}
	return lh
}

func (lh *RealBatchBuilder) Start() {
	lh.flusher.Start()
}

func (lh *RealBatchBuilder) HandleLine(line string) error {
	select {
	case lh.buffer <- line:
		return nil
	default:
		atomic.AddInt64(&lh.failures, 1)
		return fmt.Errorf("buffer full, dropping line: %s", line)
	}
}

func minInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func (lh *RealBatchBuilder) flush() error {
	lh.mtx.Lock()
	defer lh.mtx.Unlock()
	bufLen := len(lh.buffer)
	if bufLen > 0 {
		size := minInt(bufLen, lh.BatchSize)
		lines := make([]string, size)
		for i := 0; i < size; i++ {
			lines[i] = <-lh.buffer
		}
		return lh.report(lines)
	}
	return nil
}

func (lh *RealBatchBuilder) FlushWithThrottling() error {
	if time.Now().Before(lh.resumeAt) {
		log.Println("attempting to flush, but flushing is currently throttled by the server")
		log.Printf("sleeping until: %s\n", lh.resumeAt.Format(time.RFC3339))
		time.Sleep(time.Until(lh.resumeAt))
	}
	return lh.Flush()
}

func (lh *RealBatchBuilder) Flush() error {
	flushErr := lh.flush()
	if flushErr == errThrottled && lh.throttleOnBackpressure {
		atomic.AddInt64(&lh.throttled, 1)
		log.Printf("pausing requests for %v, buffer size: %d\n", lh.throttledSleepDuration, len(lh.buffer))
		lh.resumeAt = time.Now().Add(lh.throttledSleepDuration)
	}
	return flushErr
}

func (lh *RealBatchBuilder) FlushAll() error {
	lh.mtx.Lock()
	defer lh.mtx.Unlock()
	bufLen := len(lh.buffer)
	if bufLen > 0 {
		var imod int
		size := minInt(bufLen, lh.BatchSize)
		lines := make([]string, size)
		for i := 0; i < bufLen; i++ {
			imod = i % size
			lines[imod] = <-lh.buffer
			if imod == size-1 { // report batch
				if err := lh.report(lines); err != nil {
					return err
				}
			}
		}
		if imod < size-1 { // report remaining
			return lh.report(lines[0 : imod+1])
		}
	}
	return nil
}

func (lh *RealBatchBuilder) report(lines []string) error {
	strLines := strings.Join(lines, "")
	resp, err := lh.Reporter.Report(lh.format, strLines)

	if err != nil {
		if shouldRetry(err) {
			lh.bufferLines(lines)
		}
		return fmt.Errorf("error reporting %s format data to Wavefront: %q", lh.format, err)
	}

	if 400 <= resp.StatusCode && resp.StatusCode <= 599 {
		atomic.AddInt64(&lh.failures, 1)
		lh.bufferLines(lines)
		if resp.StatusCode == 406 {
			return errThrottled
		}
		return fmt.Errorf("error reporting %s format data to Wavefront. status=%d", lh.format, resp.StatusCode)
	}
	return nil
}

func shouldRetry(err error) bool {
	switch err.(type) {
	case *auth.Err:
		return false
	}
	return true
}

func (lh *RealBatchBuilder) bufferLines(batch []string) {
	log.Println("error reporting to Wavefront. buffering lines.")
	for _, line := range batch {
		_ = lh.HandleLine(line)
	}
}

func (lh *RealBatchBuilder) GetFailureCount() int64 {
	return atomic.LoadInt64(&lh.failures)
}

// GetThrottledCount returns the number of Throttled errors received.
func (lh *RealBatchBuilder) GetThrottledCount() int64 {
	return atomic.LoadInt64(&lh.throttled)
}

func (lh *RealBatchBuilder) Stop() {
	lh.flusher.Stop()
	if err := lh.FlushAll(); err != nil {
		log.Println(err)
	}
	lh.buffer = nil
}
