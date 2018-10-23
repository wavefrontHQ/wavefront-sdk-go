package senders

import (
	"fmt"
	"github.com/wavefronthq/wavefront-sdk-go/internal"
)

type proxySender struct {
	metricHandler internal.ConnectionHandler
	histoHandler  internal.ConnectionHandler
	spanHandler   internal.ConnectionHandler
	defaultSource string
}

func NewProxySender(cfg *ProxyConfiguration) (Sender, error) {
	var metricHandler internal.ConnectionHandler
	if cfg.MetricsPort != 0 {
		metricHandler = internal.NewProxyConnectionHandler(fmt.Sprintf("%s:%d", cfg.Host, cfg.MetricsPort))
	}

	var histoHandler internal.ConnectionHandler
	if cfg.DistributionPort != 0 {
		histoHandler = internal.NewProxyConnectionHandler(fmt.Sprintf("%s:%d", cfg.Host, cfg.DistributionPort))
	}

	var spanHandler internal.ConnectionHandler
	if cfg.TracingPort != 0 {
		spanHandler = internal.NewProxyConnectionHandler(fmt.Sprintf("%s:%d", cfg.Host, cfg.TracingPort))
	}

	if metricHandler == nil && histoHandler == nil && spanHandler == nil {
		return nil, fmt.Errorf("at least one proxy port should be enabled")
	}

	return &proxySender{
		defaultSource: internal.GetHostname("wavefront_proxy_sender"),
		metricHandler: metricHandler,
		histoHandler:  histoHandler,
		spanHandler:   spanHandler,
	}, nil
}

func (sender *proxySender) SendMetric(name string, value float64, ts int64, source string, tags map[string]string) error {
	if sender.metricHandler == nil {
		return fmt.Errorf("proxy metrics port not provided, cannot send metric data")
	}

	if !sender.metricHandler.Connected() {
		err := sender.metricHandler.Connect()
		if err != nil {
			return err
		}
	}

	line, err := metricLine(name, value, ts, source, tags, sender.defaultSource)
	if err != nil {
		return err
	}
	err = sender.metricHandler.SendData(line)
	return err
}

func (sender *proxySender) SendDeltaCounter(name string, value float64, source string, tags map[string]string) error {
	if name == "" {
		return fmt.Errorf("empty metric name")
	}
	if !internal.HasDeltaPrefix(name) {
		name = internal.DeltaCounterName(name)
	}
	return sender.SendMetric(name, value, 0, source, tags)
}

func (sender *proxySender) SendDistribution(name string, centroids []Centroid, hgs map[HistogramGranularity]bool, ts int64, source string, tags map[string]string) error {
	if sender.histoHandler == nil {
		return fmt.Errorf("proxy distribution port not provided, cannot send distribution data")
	}

	if !sender.histoHandler.Connected() {
		err := sender.histoHandler.Connect()
		if err != nil {
			return err
		}
	}

	line, err := histoLine(name, centroids, hgs, ts, source, tags, sender.defaultSource)
	if err != nil {
		return err
	}
	err = sender.histoHandler.SendData(line)
	return err
}

func (sender *proxySender) SendSpan(name string, startMillis, durationMillis int64, source, traceId, spanId string, parents, followsFrom []string, tags []SpanTag, spanLogs []SpanLog) error {
	if sender.spanHandler == nil {
		return fmt.Errorf("proxy tracing port not provided, cannot send span data")
	}

	if !sender.spanHandler.Connected() {
		err := sender.spanHandler.Connect()
		if err != nil {
			return err
		}
	}

	line, err := spanLine(name, startMillis, durationMillis, source, traceId, spanId, parents, followsFrom, tags, spanLogs, sender.defaultSource)
	if err != nil {
		return err
	}
	err = sender.spanHandler.SendData(line)
	return err
}

func (sender *proxySender) Close() {
	if sender.metricHandler != nil {
		sender.metricHandler.Close()
	}
	if sender.histoHandler != nil {
		sender.histoHandler.Close()
	}
	if sender.spanHandler != nil {
		sender.spanHandler.Close()
	}
}

func (sender *proxySender) Flush() error {
	errStr := ""
	if sender.metricHandler != nil {
		err := sender.metricHandler.Flush()
		if err != nil {
			errStr = errStr + err.Error() + "\n"
		}
	}
	if sender.histoHandler != nil {
		err := sender.histoHandler.Flush()
		if err != nil {
			errStr = errStr + err.Error() + "\n"
		}
	}
	if sender.spanHandler != nil {
		err := sender.spanHandler.Flush()
		if err != nil {
			errStr = errStr + err.Error()
		}
	}
	if errStr != "" {
		return fmt.Errorf(errStr)
	}
	return nil
}

func (sender *proxySender) GetFailureCount() int64 {
	var failures int64
	if sender.metricHandler != nil {
		failures += sender.metricHandler.GetFailureCount()
	}
	if sender.histoHandler != nil {
		failures += sender.histoHandler.GetFailureCount()
	}
	if sender.histoHandler != nil {
		failures += sender.histoHandler.GetFailureCount()
	}
	return failures
}
