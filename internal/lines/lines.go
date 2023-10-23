package lines

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wavefronthq/wavefront-sdk-go/internal/auth"
	"github.com/wavefronthq/wavefront-sdk-go/internal/sdkmetrics"
)

type RealHandler struct {
	// keep these two fields as first element of struct
	// to guarantee 64-bit alignment on 32-bit machines.
	// atomic.* functions crash if operands are not 64-bit aligned.
	// See https://github.com/golang/go/issues/599
	failures  int64
	throttled int64

	Reporter           Reporter
	BatchSize          int
	Format             string
	internalRegistry   sdkmetrics.Registry
	prefix             string
	lockOnErrThrottled bool
	flusher            Flusher
	BufferSize         int
	buffer             chan string
	mtx                *sync.Mutex
}


func NewLineHandler(reporter Reporter, setters ...LineHandlerOption) *RealHandler {
	handler := &RealHandler{
		Reporter:           reporter,
		// TODO: in practice is this always the same value? can we eliminate the boolean?
		lockOnErrThrottled: false,
	}

	for _, setter := range setters {
		setter(handler)
	}

	// TODO: inappropriate intimacy? Should the internal registry be creating this itself?
	if handler.internalRegistry != nil {
		handler.internalRegistry.NewGauge(handler.prefix+".queue.size", func() int64 {
			return int64(len(handler.buffer))
		})
		handler.internalRegistry.NewGauge(handler.prefix+".queue.remaining_capacity", func() int64 {
			return int64(handler.BufferSize - len(handler.buffer))
		})
	}
	return handler
}

func (h *RealHandler) Start() {
	h.flusher.Start()
}

func (h *RealHandler) HandleLine(line string) error {
	select {
	case h.buffer <- line:
		return nil
	default:
		atomic.AddInt64(&h.failures, 1)
		return fmt.Errorf("buffer full, dropping line: %s", line)
	}
}

func minInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func (h *RealHandler) StartFlusher(interval time.Duration) Flusher {
	h.flusher = NewFlusherWithBackgroundTicker(h.buffer, interval, h)
	h.flusher.Start()
	return h.flusher
}

func (h *RealHandler) Flush() error {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	bufLen := len(h.buffer)
	if bufLen > 0 {
		size := minInt(bufLen, h.BatchSize)
		lines := make([]string, size)
		for i := 0; i < size; i++ {
			lines[i] = <-h.buffer
		}
		return h.report(lines)
	}
	return nil
}

// TODO: can probably implement this in terms of Flush
func (h *RealHandler) FlushAll() error {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	bufLen := len(h.buffer)
	if bufLen > 0 {
		var imod int
		size := minInt(bufLen, h.BatchSize)
		lines := make([]string, size)
		for i := 0; i < bufLen; i++ {
			imod = i % size
			lines[imod] = <-h.buffer
			if imod == size-1 { // report batch
				if err := h.report(lines); err != nil {
					return err
				}
			}
		}
		if imod < size-1 { // report remaining
			return h.report(lines[0 : imod+1])
		}
	}
	return nil
}

func (h *RealHandler) report(lines []string) error {
	strLines := strings.Join(lines, "")
	resp, err := h.Reporter.Report(h.Format, strLines)

	if err != nil {
		if shouldRetry(err) {
			h.bufferLines(lines)
		}
		return fmt.Errorf("error reporting %s format data to Wavefront: %q", h.Format, err)
	}

	if 400 <= resp.StatusCode && resp.StatusCode <= 599 {
		atomic.AddInt64(&h.failures, 1)
		h.bufferLines(lines)
		if resp.StatusCode == 406 {
			return errThrottled
		}
		return fmt.Errorf("error reporting %s format data to Wavefront. status=%d", h.Format, resp.StatusCode)
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

func (h *RealHandler) bufferLines(batch []string) {
	log.Println("error reporting to Wavefront. buffering lines.")
	for _, line := range batch {
		_ = h.HandleLine(line)
	}
}

func (h *RealHandler) GetFailureCount() int64 {
	return atomic.LoadInt64(&h.failures)
}

// GetThrottledCount returns the number of Throttled errors received.
func (h *RealHandler) GetThrottledCount() int64 {
	return atomic.LoadInt64(&h.throttled)
}

func (h *RealHandler) Stop() {
	if h.flusher != nil {
		h.flusher.Stop()
	}
	if err := h.FlushAll(); err != nil {
		log.Println(err)
	}
}
