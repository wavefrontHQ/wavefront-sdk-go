package wavefront

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/wavefronthq/wavefront-sdk-go/event"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
	"github.com/wavefronthq/wavefront-sdk-go/internal"
	"github.com/wavefronthq/wavefront-sdk-go/types"
)

// Client Interface for sending metrics, distributions and spans to Wavefront
type Client interface {
	types.MetricSender
	types.DistributionSender
	types.SpanSender
	types.EventSender
	internal.Flusher
	Close()
}

type client struct {
	reporter         internal.Reporter
	defaultSource    string
	pointHandler     *internal.LineHandler
	histoHandler     *internal.LineHandler
	spanHandler      *internal.LineHandler
	spanLogHandler   *internal.LineHandler
	eventHandler     *internal.LineHandler
	internalRegistry *internal.MetricRegistry
}

// newWavefrontClient creates and returns a Wavefront Client instance
func newWavefrontClient(cfg *configuration) (Client, error) {
	// if cfg.Server == "" || cfg.Token == "" {
	// 	return nil, fmt.Errorf("server and token cannot be empty")
	// }
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

	client := &client{
		defaultSource: internal.GetHostname("wavefront_direct_sender"),
	}
	client.internalRegistry = internal.NewMetricRegistry(
		client,
		internal.SetPrefix("~sdk.go.core.sender.direct"),
		internal.SetTag("pid", strconv.Itoa(os.Getpid())),
	)
	client.pointHandler = makeLineHandler(reporter, cfg, internal.MetricFormat, "points", client.internalRegistry)
	client.histoHandler = makeLineHandler(reporter, cfg, internal.HistogramFormat, "histograms", client.internalRegistry)
	client.spanHandler = makeLineHandler(reporter, cfg, internal.TraceFormat, "spans", client.internalRegistry)
	client.spanLogHandler = makeLineHandler(reporter, cfg, internal.SpanLogsFormat, "span_logs", client.internalRegistry)
	client.eventHandler = makeLineHandler(reporter, cfg, internal.EventFormat, "events", client.internalRegistry)

	client.Start()
	return client, nil
}

func makeLineHandler(reporter internal.Reporter, cfg *configuration, format, prefix string,
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

func (client *client) Start() {
	client.pointHandler.Start()
	client.histoHandler.Start()
	client.spanHandler.Start()
	client.spanLogHandler.Start()
	client.internalRegistry.Start()
	client.eventHandler.Start()
}

func (client *client) SendMetric(name string, value float64, ts int64, source string, tags map[string]string) error {
	line, err := internal.MetricLine(name, value, ts, source, tags, client.defaultSource)
	if err != nil {
		return err
	}
	return client.pointHandler.HandleLine(line)
}

func (client *client) SendDeltaCounter(name string, value float64, source string, tags map[string]string) error {
	if name == "" {
		return fmt.Errorf("empty metric name")
	}
	if !internal.HasDeltaPrefix(name) {
		name = internal.DeltaCounterName(name)
	}
	if value > 0 {
		return client.SendMetric(name, value, 0, source, tags)
	}
	return nil
}

func (client *client) SendDistribution(name string, centroids []histogram.Centroid,
	hgs map[histogram.Granularity]bool, ts int64, source string, tags map[string]string) error {
	line, err := internal.HistoLine(name, centroids, hgs, ts, source, tags, client.defaultSource)
	if err != nil {
		return err
	}
	return client.histoHandler.HandleLine(line)
}

func (client *client) SendSpan(name string, startMillis, durationMillis int64, source, traceID, spanID string,
	parents, followsFrom []string, tags []types.SpanTag, spanLogs []types.SpanLog) error {
	line, err := internal.SpanLine(name, startMillis, durationMillis, source, traceID, spanID, parents, followsFrom, tags, spanLogs, client.defaultSource)
	if err != nil {
		return err
	}
	err = client.spanHandler.HandleLine(line)
	if err != nil {
		return err
	}

	if len(spanLogs) > 0 {
		logs, err := internal.SpanLogJSON(traceID, spanID, spanLogs)
		if err != nil {
			return err
		}
		return client.spanLogHandler.HandleLine(logs)
	}
	return nil
}

func (client *client) SendEvent(name string, startMillis, endMillis int64, source string, tags map[string]string, setters ...event.Option) error {
	line, err := internal.EventLineJSON(name, startMillis, endMillis, source, tags, setters...)
	if err != nil {
		return err
	}
	return client.eventHandler.HandleLine(line)
}

func (client *client) Close() {
	client.pointHandler.Stop()
	client.histoHandler.Stop()
	client.spanHandler.Stop()
	client.spanLogHandler.Stop()
	client.internalRegistry.Stop()
	client.eventHandler.Stop()
}

func (client *client) Flush() error {
	errStr := ""
	err := client.pointHandler.Flush()
	if err != nil {
		errStr = errStr + err.Error() + "\n"
	}
	err = client.histoHandler.Flush()
	if err != nil {
		errStr = errStr + err.Error() + "\n"
	}
	err = client.spanHandler.Flush()
	if err != nil {
		errStr = errStr + err.Error()
	}
	err = client.spanLogHandler.Flush()
	if err != nil {
		errStr = errStr + err.Error()
	}
	err = client.eventHandler.Flush()
	if err != nil {
		errStr = errStr + err.Error()
	}
	if errStr != "" {
		return fmt.Errorf(errStr)
	}
	return nil
}

func (client *client) GetFailureCount() int64 {
	return client.pointHandler.GetFailureCount() +
		client.histoHandler.GetFailureCount() +
		client.spanHandler.GetFailureCount() +
		client.spanLogHandler.GetFailureCount() +
		client.eventHandler.GetFailureCount()
}
