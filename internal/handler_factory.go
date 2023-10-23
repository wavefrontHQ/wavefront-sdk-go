package internal

import (
	"time"

	"github.com/wavefronthq/wavefront-sdk-go/internal/sdkmetrics"
)

type HandlerFactory struct {
	metricsReporter    Reporter
	tracesReporter     Reporter
	flushInterval      time.Duration
	bufferSize         int
	lineHandlerOptions []LineHandlerOption
}

func NewHandlerFactory(
	metricsReporter,
	tracesReporter Reporter,
	flushInterval time.Duration,
	bufferSize int,
	registry sdkmetrics.Registry) *HandlerFactory {
	return &HandlerFactory{
		metricsReporter: metricsReporter,
		tracesReporter:  tracesReporter,
		flushInterval:   flushInterval,
		bufferSize:      bufferSize,
		lineHandlerOptions: []LineHandlerOption{
			SetRegistry(registry),
		},
	}
}

func (f *HandlerFactory) NewPointHandler(batchSize int) *RealLineHandler {
	return NewLineHandler(
		f.metricsReporter,
		metricFormat,
		f.flushInterval,
		batchSize,
		f.bufferSize,
		append(f.lineHandlerOptions,
			SetHandlerPrefix("points"))...,
	)
}

func (f *HandlerFactory) NewHistogramHandler(batchSize int) *RealLineHandler {
	return NewLineHandler(
		f.metricsReporter,
		histogramFormat,
		f.flushInterval,
		batchSize,
		f.bufferSize,
		append(f.lineHandlerOptions,
			SetHandlerPrefix("histograms"))...,
	)
}

func (f *HandlerFactory) NewSpanHandler(batchSize int) *RealLineHandler {
	return NewLineHandler(
		f.tracesReporter,
		traceFormat,
		f.flushInterval,
		batchSize,
		f.bufferSize,
		append(f.lineHandlerOptions,
			SetHandlerPrefix("spans"))...,
	)
}

func (f *HandlerFactory) NewSpanLogHandler(batchSize int) *RealLineHandler {
	return NewLineHandler(
		f.tracesReporter,
		spanLogsFormat,
		f.flushInterval,
		batchSize,
		f.bufferSize,
		append(f.lineHandlerOptions,
			SetHandlerPrefix("span_logs"))...,
	)
}

// NewEventHandler creates a RealLineHandler for the Event type
// The Event handler always sets "SetLockOnThrottledError" to true
// And always uses a batch size of exactly 1.
func (f *HandlerFactory) NewEventHandler() *RealLineHandler {
	return NewLineHandler(
		f.metricsReporter,
		eventFormat,
		f.flushInterval,
		1,
		f.bufferSize,
		append(f.lineHandlerOptions,
			SetHandlerPrefix("events"),
			SetLockOnThrottledError(true))...,
	)
}
