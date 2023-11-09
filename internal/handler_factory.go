package internal

import (
	"time"

	"github.com/wavefronthq/wavefront-sdk-go/internal/sdkmetrics"
)

type SenderFactory struct {
	metricsReporter    Reporter
	tracesReporter     Reporter
	flushInterval      time.Duration
	bufferSize         int
	lineHandlerOptions []BatchAccumulatorOption
	registry           sdkmetrics.Registry
}

func NewSenderFactory(
	metricsReporter,
	tracesReporter Reporter,
	flushInterval time.Duration,
	bufferSize int,
	registry sdkmetrics.Registry) *SenderFactory {
	return &SenderFactory{
		registry:        registry,
		metricsReporter: metricsReporter,
		tracesReporter:  tracesReporter,
		flushInterval:   flushInterval,
		bufferSize:      bufferSize,
		lineHandlerOptions: []BatchAccumulatorOption{
			SetRegistry(registry),
		},
	}
}

func (f *SenderFactory) NewPointSender(batchSize int) TypedSender {
	return NewTypedSender(f.registry.PointsTracker(), f.newPointHandler(batchSize))
}

func (f *SenderFactory) NewHistogramSender(batchSize int) TypedSender {
	return NewTypedSender(f.registry.HistogramsTracker(), f.NewHistogramHandler(batchSize))
}

func (f *SenderFactory) NewSpanSender(batchSize int) TypedSender {
	return NewTypedSender(f.registry.SpansTracker(), f.newSpanHandler(batchSize))
}

func (f *SenderFactory) NewEventsSender() TypedSender {
	return NewTypedSender(f.registry.EventsTracker(), f.newEventHandler())
}

func (f *SenderFactory) NewSpanLogSender(batchSize int) TypedSender {
	return NewTypedSender(f.registry.SpanLogsTracker(), f.newSpanLogHandler(batchSize))
}

func (f *SenderFactory) newPointHandler(batchSize int) *RealBatchBuilder {
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

func (f *SenderFactory) NewHistogramHandler(batchSize int) *RealBatchBuilder {
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

func (f *SenderFactory) newSpanHandler(batchSize int) *RealBatchBuilder {
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

func (f *SenderFactory) newSpanLogHandler(batchSize int) *RealBatchBuilder {
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

// NewEventHandler creates a RealBatchBuilder for the Event type
// The Event handler always sets "ThrottleRequestsOnBackpressure" to true
// And always uses a batch size of exactly 1.
func (f *SenderFactory) newEventHandler() *RealBatchBuilder {
	return NewLineHandler(
		f.metricsReporter,
		eventFormat,
		f.flushInterval,
		1,
		f.bufferSize,
		append(f.lineHandlerOptions,
			SetHandlerPrefix("events"),
			ThrottleRequestsOnBackpressure())...,
	)
}
