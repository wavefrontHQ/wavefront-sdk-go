package senders

import (
	"fmt"
	"os"
	"strconv"

	"github.com/wavefronthq/wavefront-sdk-go/event"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
	"github.com/wavefronthq/wavefront-sdk-go/internal"
	"github.com/wavefronthq/wavefront-sdk-go/internal/delta"
	eventInternal "github.com/wavefronthq/wavefront-sdk-go/internal/event"
	histogramInternal "github.com/wavefronthq/wavefront-sdk-go/internal/histogram"
	"github.com/wavefronthq/wavefront-sdk-go/internal/metric"
	"github.com/wavefronthq/wavefront-sdk-go/internal/sdkmetrics"
	"github.com/wavefronthq/wavefront-sdk-go/internal/span"
	"github.com/wavefronthq/wavefront-sdk-go/version"
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

type realSender struct {
	defaultSource    string
	pointSender      internal.TypedSender
	histoSender      internal.TypedSender
	spanSender       internal.TypedSender
	spanLogSender    internal.TypedSender
	eventSender      internal.TypedSender
	internalRegistry sdkmetrics.Registry
	proxy            bool
}

func (sender *realSender) Start() {
	sender.internalRegistry.Start()
	sender.pointSender.Start()
	sender.histoSender.Start()
	sender.spanSender.Start()
	sender.spanLogSender.Start()
	sender.eventSender.Start()
}

func (sender *realSender) private() {
}

func (sender *realSender) SendMetric(name string, value float64, ts int64, source string, tags map[string]string) error {
	return sender.pointSender.TrySend(metric.Line(name, value, ts, source, tags, sender.defaultSource))
}

func (sender *realSender) SendDeltaCounter(name string, value float64, source string, tags map[string]string) error {
	if value > 0 {
		return sender.pointSender.TrySend(delta.Line(name, value, source, tags, sender.defaultSource))
	}
	return nil
}

func (sender *realSender) SendDistribution(
	name string,
	centroids []histogram.Centroid,
	hgs map[histogram.Granularity]bool,
	ts int64,
	source string,
	tags map[string]string,
) error {
	return sender.histoSender.TrySend(histogramInternal.Line(name, centroids, hgs, ts, source, tags, sender.defaultSource))
}

func (sender *realSender) SendSpan(
	name string,
	startMillis, durationMillis int64,
	source, traceID, spanID string,
	parents, followsFrom []string,
	tags []SpanTag,
	spanLogs []SpanLog,
) error {
	line, err := span.Line(
		name,
		startMillis,
		durationMillis,
		source,
		traceID,
		spanID,
		parents,
		followsFrom,
		makeSpanTags(tags),
		makeSpanLogs(spanLogs),
		sender.defaultSource,
	)

	err = sender.spanSender.TrySend(line, err)
	if err != nil {
		return err
	}

	if len(spanLogs) > 0 {
		return sender.spanLogSender.TrySend(span.LogJSON(traceID, spanID, makeSpanLogs(spanLogs), line))
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

func (sender *realSender) SendEvent(
	name string,
	startMillis, endMillis int64,
	source string,
	tags map[string]string,
	setters ...event.Option,
) error {
	if sender.proxy {
		return sender.eventSender.TrySend(eventInternal.Line(name, startMillis, endMillis, source, tags, setters...))
	}
	return sender.eventSender.TrySend(eventInternal.LineJSON(name, startMillis, endMillis, source, tags, setters...))
}

func (sender *realSender) Close() {
	sender.pointSender.Stop()
	sender.histoSender.Stop()
	sender.spanSender.Stop()
	sender.spanLogSender.Stop()
	sender.internalRegistry.Stop()
	sender.eventSender.Stop()
}

func (sender *realSender) Flush() error {
	errStr := ""
	err := sender.pointSender.Flush()
	if err != nil {
		errStr = errStr + err.Error() + "\n"
	}
	err = sender.histoSender.Flush()
	if err != nil {
		errStr = errStr + err.Error() + "\n"
	}
	err = sender.spanSender.Flush()
	if err != nil {
		errStr = errStr + err.Error()
	}
	err = sender.spanLogSender.Flush()
	if err != nil {
		errStr = errStr + err.Error()
	}
	err = sender.eventSender.Flush()
	if err != nil {
		errStr = errStr + err.Error()
	}
	if errStr != "" {
		return fmt.Errorf(errStr)
	}
	return nil
}

func (sender *realSender) GetFailureCount() int64 {
	return sender.pointSender.GetFailureCount() +
		sender.histoSender.GetFailureCount() +
		sender.spanSender.GetFailureCount() +
		sender.spanLogSender.GetFailureCount() +
		sender.eventSender.GetFailureCount()
}

func (sender *realSender) realInternalRegistry(cfg *configuration) sdkmetrics.Registry {
	var setters []sdkmetrics.RegistryOption

	setters = append(setters, sdkmetrics.SetPrefix(cfg.MetricPrefix()))
	setters = append(setters, sdkmetrics.SetTag("pid", strconv.Itoa(os.Getpid())))
	setters = append(setters, sdkmetrics.SetTag("version", version.Version))

	for key, value := range cfg.SDKMetricsTags {
		setters = append(setters, sdkmetrics.SetTag(key, value))
	}

	return sdkmetrics.NewMetricRegistry(
		sender,
		setters...,
	)
}
