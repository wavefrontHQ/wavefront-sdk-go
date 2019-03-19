package internal

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type LineHandler struct {
	Reporter      Reporter
	BatchSize     int
	MaxBufferSize int
	Format        string
	flushTicker   *time.Ticker
	mtx           sync.Mutex
	failures      int64
	buffer        chan string
	done          chan struct{}
}

func NewLineHandler(reporter Reporter, format string, flushInterval time.Duration, batchSize, maxBufferSize int) *LineHandler {
	return &LineHandler{
		Reporter:      reporter,
		BatchSize:     batchSize,
		MaxBufferSize: maxBufferSize,
		flushTicker:   time.NewTicker(flushInterval),
		Format:        format,
	}
}

func (lh *LineHandler) Start() {
	lh.buffer = make(chan string, lh.MaxBufferSize)
	lh.done = make(chan struct{})

	go func() {
		for {
			select {
			case <-lh.flushTicker.C:
				err := lh.Flush()
				if err != nil {
					log.Println(err)
				}
			case <-lh.done:
				return
			}
		}
	}()
}

func (lh *LineHandler) HandleLine(line string) error {
	select {
	case lh.buffer <- line:
		return nil
	default:
		atomic.AddInt64(&lh.failures, 1)
		return fmt.Errorf("buffer full, dropping line: %s", line)
	}
}

func (lh *LineHandler) Flush() error {
	lh.mtx.Lock()
	defer lh.mtx.Unlock()
	bufLen := len(lh.buffer)
	if bufLen > 0 {
		size := min(bufLen, lh.BatchSize)
		lines := make([]string, size)
		for i := 0; i < size; i++ {
			lines[i] = <-lh.buffer
		}
		return lh.report(lines)
	}
	return nil
}

func (lh *LineHandler) FlushAll() error {
	lh.mtx.Lock()
	defer lh.mtx.Unlock()
	bufLen := len(lh.buffer)
	if bufLen > 0 {
		var imod int
		size := min(bufLen, lh.BatchSize)
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

func (lh *LineHandler) report(lines []string) error {
	strLines := strings.Join(lines, "")
	resp, err := lh.Reporter.Report(lh.Format, strLines)

	if err != nil || (400 <= resp.StatusCode && resp.StatusCode <= 599) {
		atomic.AddInt64(&lh.failures, 1)
		lh.bufferLines(lines)
		if err != nil {
			return fmt.Errorf("error reporting %s format data to Wavefront: %q", lh.Format, err)
		} else {
			return fmt.Errorf("error reporting %s format data to Wavefront. status=%d", lh.Format, resp.StatusCode)
		}
	}
	return nil
}

func (lh *LineHandler) bufferLines(batch []string) {
	log.Println("error reporting to Wavefront. buffering lines.")
	for _, line := range batch {
		lh.HandleLine(line)
	}
}

func (lh *LineHandler) GetFailureCount() int64 {
	return atomic.LoadInt64(&lh.failures)
}

func (lh *LineHandler) Stop() {
	lh.flushTicker.Stop()
	lh.done <- struct{}{} // block until goroutine exits
	if err := lh.FlushAll(); err != nil {
		log.Println(err)
	}
	lh.done = nil
	lh.buffer = nil
}
