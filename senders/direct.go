package senders

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/wavefronthq/wavefront-sdk-go/event"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
	"github.com/wavefronthq/wavefront-sdk-go/internal"
)

type directSender struct {
	reporter         internal.Reporter
	defaultSource    string
	handlers         []*internal.LineHandler
	internalRegistry *internal.MetricRegistry
}

// NewDirectSender creates and returns a Wavefront Direct Ingestion Sender instance
func NewDirectSender(cfg *DirectConfiguration) (Sender, error) {
	if cfg.Server == "" || cfg.Token == "" {
		return nil, fmt.Errorf("server and token cannot be empty")
	}
	if cfg.BatchSize == 0 {
		cfg.BatchSize = defaultBatchSize
	}
	if cfg.MaxBufferSize == 0 {
		cfg.MaxBufferSize = defaultBufferSize
	}
	if cfg.FlushIntervalSeconds == 0 {
		cfg.FlushIntervalSeconds = defaultFlushInterval
	}

	reporter := internal.NewDirectReporter(cfg.Server, cfg.Token)

	sender := &directSender{
		defaultSource: internal.GetHostname("wavefront_direct_sender"),
		handlers:      make([]*internal.LineHandler, HandlersCount),
	}
	sender.internalRegistry = internal.NewMetricRegistry(
		sender,
		internal.SetPrefix("~sdk.go.core.sender.direct"),
		internal.SetTag("pid", strconv.Itoa(os.Getpid())),
	)
	sender.handlers[MetricHandler] = makeLineHandler(reporter, cfg, internal.MetricFormat, "points", sender.internalRegistry)
	sender.handlers[HistoHandler] = makeLineHandler(reporter, cfg, internal.HistogramFormat, "histograms", sender.internalRegistry)
	sender.handlers[SpanHandler] = makeLineHandler(reporter, cfg, internal.TraceFormat, "spans", sender.internalRegistry)
	sender.handlers[SpanHandler] = makeLineHandler(reporter, cfg, internal.SpanLogsFormat, "span_logs", sender.internalRegistry)
	sender.handlers[EventHandler] = makeLineHandler(reporter, cfg, internal.EventFormat, "events", sender.internalRegistry)

	sender.Start()
	return sender, nil
}

func makeLineHandler(reporter internal.Reporter, cfg *DirectConfiguration, format, prefix string,
	registry *internal.MetricRegistry) *internal.LineHandler {
	flushInterval := time.Second * time.Duration(cfg.FlushIntervalSeconds)

	opts := []internal.LineHandlerOption{internal.SetHandlerPrefix(prefix), internal.SetRegistry(registry)}
	batchSize := cfg.BatchSize
	if format == internal.EventFormat {
		batchSize = 1
		opts = append(opts, internal.SetLockOnThrottledError(true))
	}

	return internal.NewLineHandler(reporter, format, flushInterval, batchSize, cfg.MaxBufferSize, opts...)
}

func (sender *directSender) Start() {
	for _, h := range sender.handlers {
		if h != nil {
			h.Start()
		}
	}
	sender.internalRegistry.Start()
}

func (sender *directSender) SendMetric(name string, value float64, ts int64, source string, tags map[string]string) error {
	line, err := MetricLine(name, value, ts, source, tags, sender.defaultSource)
	if err != nil {
		return err
	}
	return sender.handlers[MetricHandler].HandleLine(line)
}

func (sender *directSender) SendDeltaCounter(name string, value float64, source string, tags map[string]string) error {
	if name == "" {
		return fmt.Errorf("empty metric name")
	}
	if !internal.HasDeltaPrefix(name) {
		name = internal.DeltaCounterName(name)
	}
	return sender.SendMetric(name, value, 0, source, tags)
}

func (sender *directSender) SendDistribution(name string, centroids []histogram.Centroid,
	hgs map[histogram.Granularity]bool, ts int64, source string, tags map[string]string) error {
	line, err := HistoLine(name, centroids, hgs, ts, source, tags, sender.defaultSource)
	if err != nil {
		return err
	}
	return sender.handlers[HistoHandler].HandleLine(line)
}

func (sender *directSender) SendSpan(name string, startMillis, durationMillis int64, source, traceId, spanId string,
	parents, followsFrom []string, tags []SpanTag, spanLogs []SpanLog) error {
	line, err := SpanLine(name, startMillis, durationMillis, source, traceId, spanId, parents, followsFrom, tags, spanLogs, sender.defaultSource)
	if err != nil {
		return err
	}
	err = sender.handlers[SpanHandler].HandleLine(line)
	if err != nil {
		return err
	}

	if len(spanLogs) > 0 {
		logs, err := SpanLogJSON(traceId, spanId, spanLogs)
		if err != nil {
			return err
		}
		return sender.handlers[SpanHandler].HandleLine(logs)
	}
	return nil
}

func (sender *directSender) SendEvent(name string, startMillis, endMillis int64, source string, tags map[string]string, setters ...event.Option) error {
	line, err := EventLineJSOM(name, startMillis, endMillis, source, tags, setters...)
	if err != nil {
		return err
	}
	return sender.handlers[EventHandler].HandleLine(line)
}

func (sender *directSender) Close() {
	for _, h := range sender.handlers {
		if h != nil {
			h.Stop()
		}
	}
	sender.internalRegistry.Stop()
}

func (sender *directSender) Flush() error {
	errStr := ""
	for _, h := range sender.handlers {
		if h != nil {
			err := h.Flush()
			if err != nil {
				errStr = errStr + err.Error() + "\n"
			}
		}
	}
	if errStr != "" {
		return errors.New(strings.Trim(errStr, "\n"))
	}
	return nil
}

func (sender *directSender) GetFailureCount() int64 {
	var failures int64
	for _, h := range sender.handlers {
		if h != nil {
			failures += h.GetFailureCount()
		}
	}
	return failures
}
