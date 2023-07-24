package senders

import (
	"fmt"
	"github.com/wavefronthq/wavefront-sdk-go/event"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
	"github.com/wavefronthq/wavefront-sdk-go/internal"
	eventInternal "github.com/wavefronthq/wavefront-sdk-go/internal/event"
	histogramInternal "github.com/wavefronthq/wavefront-sdk-go/internal/histogram"
	"github.com/wavefronthq/wavefront-sdk-go/internal/metric"
	"github.com/wavefronthq/wavefront-sdk-go/internal/span"
)

// Sender Interface for sending metrics, distributions and spans to Wavefront
type Sender interface {
	MetricSender
	DistributionSender
	SpanSender
	EventSender
	internal.Flusher
	Close()
	private()
}

type wavefrontSender struct {
	reporter         internal.Reporter
	defaultSource    string
	pointHandler     *internal.LineHandler
	histoHandler     *internal.LineHandler
	spanHandler      *internal.LineHandler
	spanLogHandler   *internal.LineHandler
	eventHandler     *internal.LineHandler
	internalRegistry internal.MetricRegistry
	proxy            bool
}

func newLineHandler(reporter internal.Reporter, cfg *configuration, format, prefix string, registry internal.MetricRegistry) *internal.LineHandler {
	opts := []internal.LineHandlerOption{internal.SetHandlerPrefix(prefix), internal.SetRegistry(registry)}
	batchSize := cfg.BatchSize
	if format == internal.EventFormat {
		batchSize = 1
		opts = append(opts, internal.SetLockOnThrottledError(true))
	}

	return internal.NewLineHandler(reporter, format, cfg.FlushInterval, batchSize, cfg.MaxBufferSize, opts...)
}

func (sender *wavefrontSender) Start() {
	sender.pointHandler.Start()
	sender.histoHandler.Start()
	sender.spanHandler.Start()
	sender.spanLogHandler.Start()
	sender.internalRegistry.Start()
	sender.eventHandler.Start()
}

func (sender *wavefrontSender) private() {
}

func (sender *wavefrontSender) SendMetric(name string, value float64, ts int64, source string, tags map[string]string) error {
	line, err := metric.Line(name, value, ts, source, tags, sender.defaultSource)
	return trySendWith(
		line,
		err,
		sender.pointHandler,
		sender.internalRegistry.PointsTracker(),
	)
}

func (sender *wavefrontSender) SendDeltaCounter(name string, value float64, source string, tags map[string]string) error {
	if name == "" {
		sender.internalRegistry.PointsTracker().IncInvalid()
		return fmt.Errorf("empty metric name")
	}
	if !internal.HasDeltaPrefix(name) {
		name = internal.DeltaCounterName(name)
	}
	if value > 0 {
		return sender.SendMetric(name, value, 0, source, tags)
	}
	return nil
}

func (sender *wavefrontSender) SendDistribution(name string, centroids []histogram.Centroid,
	hgs map[histogram.Granularity]bool, ts int64, source string, tags map[string]string) error {
	line, err := histogramInternal.HistogramLine(name, centroids, hgs, ts, source, tags, sender.defaultSource)
	return trySendWith(
		line,
		err,
		sender.histoHandler,
		sender.internalRegistry.HistogramsTracker(),
	)
}

func trySendWith(line string, err error, handler *internal.LineHandler, tracker internal.SuccessTracker) error {
	if err != nil {
		tracker.IncValid()
		return err
	} else {
		tracker.IncValid()
	}
	err = handler.HandleLine(line)
	if err != nil {
		tracker.IncDropped()
	}
	return err
}

func (sender *wavefrontSender) SendSpan(name string, startMillis, durationMillis int64, source, traceId, spanId string,
	parents, followsFrom []string, tags []SpanTag, spanLogs []SpanLog) error {

	logs := makeSpanLogs(spanLogs)
	line, err := span.Line(name, startMillis, durationMillis, source, traceId, spanId, parents, followsFrom, makeSpanTags(tags), logs, sender.defaultSource)
	err = trySendWith(
		line,
		err,
		sender.spanHandler,
		sender.internalRegistry.SpansTracker())
	if err != nil {
		return err
	}

	if len(spanLogs) > 0 {
		logJSON, logJSONErr := span.LogJSON(traceId, spanId, logs, line)
		return trySendWith(
			logJSON,
			logJSONErr,
			sender.spanLogHandler,
			sender.internalRegistry.SpanLogsTracker())
	}
	return nil
}

func makeSpanTags(tags []SpanTag) []span.Tag {
	spanTags := make([]span.Tag, len(tags))
	for i, tag := range tags {
		spanTags[i] = span.Tag(tag)
	}
	return spanTags
}

func makeSpanLogs(logs []SpanLog) []span.Log {
	spanLogs := make([]span.Log, len(logs))
	for i, log := range logs {
		spanLogs[i] = span.Log(log)
	}
	return spanLogs
}

func (sender *wavefrontSender) SendEvent(name string, startMillis, endMillis int64, source string, tags map[string]string, setters ...event.Option) error {
	var line string
	var err error
	if sender.proxy {
		line, err = eventInternal.Line(name, startMillis, endMillis, source, tags, setters...)
	} else {
		line, err = eventInternal.LineJSON(name, startMillis, endMillis, source, tags, setters...)
	}

	return trySendWith(
		line,
		err,
		sender.eventHandler,
		sender.internalRegistry.EventsTracker(),
	)
}

func (sender *wavefrontSender) Close() {
	sender.pointHandler.Stop()
	sender.histoHandler.Stop()
	sender.spanHandler.Stop()
	sender.spanLogHandler.Stop()
	sender.internalRegistry.Stop()
	sender.eventHandler.Stop()
}

func (sender *wavefrontSender) Flush() error {
	errStr := ""
	err := sender.pointHandler.Flush()
	if err != nil {
		errStr = errStr + err.Error() + "\n"
	}
	err = sender.histoHandler.Flush()
	if err != nil {
		errStr = errStr + err.Error() + "\n"
	}
	err = sender.spanHandler.Flush()
	if err != nil {
		errStr = errStr + err.Error()
	}
	err = sender.spanLogHandler.Flush()
	if err != nil {
		errStr = errStr + err.Error()
	}
	err = sender.eventHandler.Flush()
	if err != nil {
		errStr = errStr + err.Error()
	}
	if errStr != "" {
		return fmt.Errorf(errStr)
	}
	return nil
}

func (sender *wavefrontSender) GetFailureCount() int64 {
	return sender.pointHandler.GetFailureCount() +
		sender.histoHandler.GetFailureCount() +
		sender.spanHandler.GetFailureCount() +
		sender.spanLogHandler.GetFailureCount() +
		sender.eventHandler.GetFailureCount()
}
