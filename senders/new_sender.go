package senders

import (
	"fmt"

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
	metricsReporter := internal.NewReporter(cfg.metricsURL(), tokenService, client)
	tracesReporter := internal.NewReporter(cfg.tracesURL(), tokenService, client)

	sender := &realSender{
		defaultSource: internal.GetHostname("wavefront_direct_sender"),
		proxy:         !cfg.Direct(),
	}
	if cfg.SendInternalMetrics {
		sender.internalRegistry = sender.realInternalRegistry(cfg)
	} else {
		sender.internalRegistry = sdkmetrics.NewNoOpRegistry()
	}

	hf := internal.NewHandlerFactory(
		metricsReporter,
		tracesReporter,
		cfg.FlushInterval,
		cfg.MaxBufferSize,
		sender.internalRegistry,
	)

	sender.pointHandler = hf.NewPointHandler(cfg.BatchSize)
	sender.histoHandler = hf.NewHistogramHandler(cfg.BatchSize)
	sender.spanHandler = hf.NewSpanHandler(cfg.BatchSize)
	sender.spanLogHandler = hf.NewSpanLogHandler(cfg.BatchSize)
	sender.eventHandler = hf.NewEventHandler()
	sender.Start()
	return sender, nil
}
