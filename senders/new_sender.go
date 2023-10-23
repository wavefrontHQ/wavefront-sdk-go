package senders

import (
	"fmt"
	"github.com/wavefronthq/wavefront-sdk-go/internal/lines"

	"github.com/wavefronthq/wavefront-sdk-go/internal"
	"github.com/wavefronthq/wavefront-sdk-go/internal/sdkmetrics"
)

// NewSender creates a Sender using the provided URL and Options
func NewSender(wfURL string, setters ...Option) (Sender, error) {
	cfg, err := createConfig(wfURL, setters...)
	if err != nil {
		return nil, fmt.Errorf("unable to create sender config: %s", err)
	}

	tokenService := tokenServiceForCfg(cfg)
	client := cfg.HTTPClient
	metricsReporter := lines.NewReporter(cfg.metricsURL(), tokenService, client)
	tracesReporter := lines.NewReporter(cfg.tracesURL(), tokenService, client)

	sender := &realSender{
		defaultSource: internal.GetHostname("wavefront_direct_sender"),
		proxy:         !cfg.Direct(),
	}
	if cfg.SendInternalMetrics {
		sender.internalRegistry = sender.realInternalRegistry(cfg)
	} else {
		sender.internalRegistry = sdkmetrics.NewNoOpRegistry()
	}

	//TODO: style: could have format-specific handler factories instead
	// this would avoid exposing the format string constants and the prefixes

	factory := lines.NewHandlerFactory(metricsURL, tracesURL, tokenService, client, internalRegistry)
	factory.NewPointsHandler()
	factory.NewHistogramHandler()

	// BatchSender
	// uses a ReportHTTPRequestSender
	// can optionally use a BackgroundFlusher

	sender.pointHandler = lines.NewHandler(metricsReporter, cfg, lines.MetricFormat, "points", sender.internalRegistry)
	sender.histoHandler = lines.NewHandler(metricsReporter, cfg, lines.HistogramFormat, "histograms", sender.internalRegistry)
	sender.spanHandler = lines.NewHandler(tracesReporter, cfg, lines.TraceFormat, "spans", sender.internalRegistry)
	sender.spanLogHandler = lines.NewHandler(tracesReporter, cfg, lines.SpanLogsFormat, "span_logs", sender.internalRegistry)
	sender.eventHandler = lines.NewHandler(metricsReporter, cfg, lines.EventFormat, "events", sender.internalRegistry)

	sender.Start()
	return sender, nil
}

func copyTags(orig map[string]string) map[string]string {
	result := make(map[string]string, len(orig))
	for key, value := range orig {
		result[key] = value
	}
	return result
}
