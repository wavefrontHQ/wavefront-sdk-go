package senders

import (
	"fmt"
	"time"

	"github.com/wavefronthq/wavefront-sdk-go/event"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
	"github.com/wavefronthq/wavefront-sdk-go/internal"
)

// Sender is for sending metrics, distributions, and spans to Wavefront.
// Method calls on the zero value are no-ops and always returns success.
type Sender struct {
	spec senderSpec
}

// SendMetric sends a single metric to Wavefront with optional timestamp and
// tags.
func (s *Sender) SendMetric(name string, value float64, ts int64, source string, tags map[string]string) error {
	if s.spec == nil {
		return nil
	}
	return s.spec.SendMetric(name, value, ts, source, tags)
}

// SendDeltaCounter sends a delta counter (counter aggregated at the Wavefront
// service) to Wavefront.
func (s *Sender) SendDeltaCounter(name string, value float64, source string, tags map[string]string) error {
	if s.spec == nil {
		return nil
	}
	return s.spec.SendDeltaCounter(name, value, source, tags)
}

// SendDistribution sends a distribution of metrics to Wavefront with optional
// timestamp and tags. Each centroid is a 2-dimensional entity with the first
// dimension the mean value and the second dimension the count of points in
// the centroid. The granularity informs the set of intervals (minute, hour,
// and/or day) by which the histogram data should be aggregated.
func (s *Sender) SendDistribution(name string, centroids []histogram.Centroid,
	hgs map[histogram.Granularity]bool, ts int64, source string, tags map[string]string) error {
	if s.spec == nil {
		return nil
	}
	return s.spec.SendDistribution(name, centroids, hgs, ts, source, tags)
}

// SendSpan sends a tracing span to Wavefront.  traceId, spanId, parentIds and
// preceding spanIds are expected to be UUID strings. parents and preceding
// spans can be empty for a root span. Span tag keys can be repeated (example:
// "user"="foo" and "user"="bar") span logs are currently omitted.
func (s *Sender) SendSpan(name string, startMillis, durationMillis int64, source, traceId, spanId string,
	parents, followsFrom []string, tags []SpanTag, spanLogs []SpanLog) error {
	if s.spec == nil {
		return nil
	}
	return s.spec.SendSpan(name, startMillis, durationMillis, source, traceId, spanId, parents, followsFrom, tags, spanLogs)
}

// SendEvent sends an event to Wavefront with optional tags.
func (s *Sender) SendEvent(name string, startMillis, endMillis int64, source string, tags map[string]string, setters ...event.Option) error {
	if s.spec == nil {
		return nil
	}
	return s.spec.SendEvent(name, startMillis, endMillis, source, tags, setters...)
}

// Close closes the Sender.
func (s *Sender) Close() {
	if s.spec != nil {
		s.spec.Close()
	}
}

// Start starts autoflushing in this Sender
func (s *Sender) Start() {
	if s.spec != nil {
		s.spec.Start()
	}
}

// Flush manually flushes this Sender.
func (s *Sender) Flush() error {
	if s.spec == nil {
		return nil
	}
	return s.spec.Flush()
}

// GetFailureCount returns the number of cumulative failures for this Sender
func (s *Sender) GetFailureCount() int64 {
	if s.spec == nil {
		return 0
	}
	return s.spec.GetFailureCount()
}

// Sender Interface for sending metrics, distributions and spans to Wavefront
type senderSpec interface {
	internal.Flusher
	Close()
	SendEvent(name string, millis int64, millis2 int64, source string, tags map[string]string, setters ...event.Option) error
	SendSpan(name string, millis int64, millis2 int64, source string, id string, id2 string, parents []string, from []string, tags []SpanTag, logs []SpanLog) error
	SendDistribution(name string, centroids []histogram.Centroid, hgs map[histogram.Granularity]bool, ts int64, source string, tags map[string]string) error
	SendDeltaCounter(name string, value float64, source string, tags map[string]string) error
	SendMetric(name string, value float64, ts int64, source string, tags map[string]string) error
}

type wavefrontSender struct {
	reporter         internal.Reporter
	defaultSource    string
	pointHandler     *internal.LineHandler
	histoHandler     *internal.LineHandler
	spanHandler      *internal.LineHandler
	spanLogHandler   *internal.LineHandler
	eventHandler     *internal.LineHandler
	internalRegistry *internal.MetricRegistry

	pointsValid   *internal.DeltaCounter
	pointsInvalid *internal.DeltaCounter
	pointsDropped *internal.DeltaCounter

	histogramsValid   *internal.DeltaCounter
	histogramsInvalid *internal.DeltaCounter
	histogramsDropped *internal.DeltaCounter

	spansValid   *internal.DeltaCounter
	spansInvalid *internal.DeltaCounter
	spansDropped *internal.DeltaCounter

	spanLogsValid   *internal.DeltaCounter
	spanLogsInvalid *internal.DeltaCounter
	spanLogsDropped *internal.DeltaCounter

	eventsValid   *internal.DeltaCounter
	eventsInvalid *internal.DeltaCounter
	eventsDropped *internal.DeltaCounter

	proxy bool
}

func newLineHandler(reporter internal.Reporter, cfg *configuration, format, prefix string, registry *internal.MetricRegistry) *internal.LineHandler {
	flushInterval := time.Second * time.Duration(cfg.FlushIntervalSeconds)

	opts := []internal.LineHandlerOption{internal.SetHandlerPrefix(prefix), internal.SetRegistry(registry)}
	batchSize := cfg.BatchSize
	if format == internal.EventFormat {
		batchSize = 1
		opts = append(opts, internal.SetLockOnThrottledError(true))
	}

	return internal.NewLineHandler(reporter, format, flushInterval, batchSize, cfg.MaxBufferSize, opts...)
}

func (sender *wavefrontSender) Start() {
	sender.pointHandler.Start()
	sender.histoHandler.Start()
	sender.spanHandler.Start()
	sender.spanLogHandler.Start()
	sender.internalRegistry.Start()
	sender.eventHandler.Start()
}

func (sender *wavefrontSender) SendMetric(name string, value float64, ts int64, source string, tags map[string]string) error {
	line, err := metricLine(name, value, ts, source, tags, sender.defaultSource)
	if err != nil {
		sender.pointsInvalid.Inc()
		return err
	} else {
		sender.pointsValid.Inc()
	}
	err = sender.pointHandler.HandleLine(line)
	if err != nil {
		sender.pointsDropped.Inc()
	}
	return err
}
func (sender *wavefrontSender) SendDeltaCounter(name string, value float64, source string, tags map[string]string) error {
	if name == "" {
		sender.pointsInvalid.Inc()
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
	line, err := histogramLine(name, centroids, hgs, ts, source, tags, sender.defaultSource)
	if err != nil {
		sender.histogramsInvalid.Inc()
		return err
	} else {
		sender.histogramsValid.Inc()
	}
	err = sender.histoHandler.HandleLine(line)
	if err != nil {
		sender.histogramsDropped.Inc()
	}
	return err
}
func (sender *wavefrontSender) SendSpan(name string, startMillis, durationMillis int64, source, traceId, spanId string,
	parents, followsFrom []string, tags []SpanTag, spanLogs []SpanLog) error {
	line, err := spanLine(name, startMillis, durationMillis, source, traceId, spanId, parents, followsFrom, tags, spanLogs, sender.defaultSource)
	if err != nil {
		sender.spansInvalid.Inc()
		return err
	} else {
		sender.spansValid.Inc()
	}
	err = sender.spanHandler.HandleLine(line)
	if err != nil {
		sender.spansDropped.Inc()
		return err
	}

	if len(spanLogs) > 0 {
		logs, err := spanLogJSON(traceId, spanId, spanLogs, line)
		if err != nil {
			sender.spanLogsInvalid.Inc()
			return err
		} else {
			sender.spanLogsValid.Inc()
		}
		err = sender.spanLogHandler.HandleLine(logs)
		if err != nil {
			sender.spanLogsDropped.Inc()
		}
		return err
	}
	return nil
}

func (sender *wavefrontSender) SendEvent(name string, startMillis, endMillis int64, source string, tags map[string]string, setters ...event.Option) error {
	var line string
	var err error
	if sender.proxy {
		line, err = eventLine(name, startMillis, endMillis, source, tags, setters...)
	} else {
		line, err = eventLineJSON(name, startMillis, endMillis, source, tags, setters...)
	}
	if err != nil {
		sender.eventsInvalid.Inc()
		return err
	} else {
		sender.eventsValid.Inc()
	}
	err = sender.eventHandler.HandleLine(line)
	if err != nil {
		sender.eventsDropped.Inc()
	}
	return err
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
