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
	FlushTicker   *time.Ticker
	Format        string
	failures      int64
	mtx           sync.Mutex
	lines         []string
}

func (lh *LineHandler) Start() {
	go func() {
		for range lh.FlushTicker.C {
			err := lh.Flush()
			if err != nil {
				log.Println(err)
			}
		}
	}()
}

func (lh *LineHandler) HandleLine(line string) error {
	lh.mtx.Lock()
	defer lh.mtx.Unlock()

	if len(lh.lines) >= lh.MaxBufferSize {
		atomic.AddInt64(&lh.failures, 1)
		return fmt.Errorf("buffer full, dropping line: %s", line)
	}
	lh.lines = append(lh.lines, line)
	return nil
}

func (lh *LineHandler) Flush() error {
	batch := lh.linesBatch()
	if len(batch) == 0 {
		return nil
	}
	batchStr := strings.Join(batch, "")
	resp, err := lh.Reporter.Report(lh.Format, batchStr)

	if err != nil || (400 <= resp.StatusCode && resp.StatusCode <= 599) {
		atomic.AddInt64(&lh.failures, 1)
		lh.bufferLines(batch)
		if err != nil {
			return fmt.Errorf("error reporting %s format data to Wavefront: %q", lh.Format, err)
		} else {
			return fmt.Errorf("error reporting %s format data to Wavefront. status=%d", lh.Format, resp.StatusCode)
		}
	}
	return nil
}

func (lh *LineHandler) linesBatch() []string {
	lh.mtx.Lock()
	defer lh.mtx.Unlock()

	currLen := len(lh.lines)
	batchSize := min(currLen, lh.BatchSize)
	batchLines := lh.lines[:batchSize]
	lh.lines = lh.lines[batchSize:currLen]
	return batchLines
}

func (lh *LineHandler) bufferLines(batch []string) {
	lh.mtx.Lock()
	defer lh.mtx.Unlock()

	currLen := len(lh.lines)
	remaining := lh.MaxBufferSize - currLen
	if remaining > 0 {
		drainLen := min(remaining, len(batch))
		drainLines := batch[:drainLen]
		lh.lines = append(lh.lines, drainLines...)
	}
}

func (lh *LineHandler) GetFailureCount() int64 {
	return atomic.LoadInt64(&lh.failures)
}

func (lh *LineHandler) Stop() {
	lh.FlushTicker.Stop()
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
