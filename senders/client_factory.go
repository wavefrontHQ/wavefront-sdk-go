package senders

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/wavefronthq/wavefront-sdk-go/internal"
	"github.com/wavefronthq/wavefront-sdk-go/version"
)

const (
	defaultTracesPort              = 30001
	defaultMetricsPort             = 2878
	defaultBatchSize               = 10_000
	defaultBufferSize              = 50_000
	defaultFlushInterval           = 1 * time.Second
	defaultInternalMetricsInterval = 60 * time.Second
	defaultTimeout                 = 10 * time.Second
)

// Option Wavefront client configuration options
type Option func(*configuration)

// Configuration for the direct ingestion sender
type configuration struct {
	Server string // Wavefront URL of the form https://<INSTANCE>.wavefront.com
	Token  string // Wavefront API token with direct data ingestion permission

	// Optional configuration properties. Default values should suffice for most use cases.
	// override the defaults only if you wish to set higher values.

	MetricsPort int
	TracesPort  int

	// max batch of data sent per flush interval. defaults to 10,000. recommended not to exceed 40,000.
	BatchSize int

	// Enable or disable internal SDK metrics that begin with ~sdk.go.core
	InternalMetricsEnabled bool

	// interval at which to report internal SDK metrics
	InternalMetricsInterval time.Duration

	// size of internal buffers beyond which received data is dropped.
	// helps with handling brief increases in data and buffering on errors.
	// separate buffers are maintained per data type (metrics, spans and distributions)
	// buffers are not pre-allocated to max size and vary based on actual usage.
	// defaults to 500,000. higher values could use more memory.
	MaxBufferSize int

	// interval (in seconds) at which to flush data to Wavefront. defaults to 1 Second.
	// together with batch size controls the max theoretical throughput of the sender.
	FlushInterval  time.Duration
	SDKMetricsTags map[string]string
	Path           string

	Timeout time.Duration

	TLSConfig *tls.Config
}

func (c *configuration) Direct() bool {
	return c.Token != ""
}

func (c *configuration) MetricPrefix() string {
	result := "~sdk.go.core.sender.proxy"
	if c.Direct() {
		result = "~sdk.go.core.sender.direct"
	}
	return result
}

func (c *configuration) setDefaultPort(port int) {
	c.MetricsPort = port
	c.TracesPort = port
}

// NewSender creates Wavefront Sender using the provided URL and Options
func NewSender(wfURL string, setters ...Option) (Sender, error) {
	cfg, err := createConfig(wfURL, setters...)
	if err != nil {
		return nil, fmt.Errorf("unable to create sender config: %s", err)
	}
	return newSender(cfg)
}

func createConfig(wfURL string, setters ...Option) (*configuration, error) {
	cfg := &configuration{
		MetricsPort:             defaultMetricsPort,
		TracesPort:              defaultTracesPort,
		BatchSize:               defaultBatchSize,
		MaxBufferSize:           defaultBufferSize,
		FlushInterval:           defaultFlushInterval,
		InternalMetricsEnabled:  true,
		InternalMetricsInterval: defaultInternalMetricsInterval,
		SDKMetricsTags:          map[string]string{},
		Timeout:                 defaultTimeout,
	}

	u, err := url.Parse(wfURL)
	if err != nil {
		return nil, err
	}

	if len(u.User.String()) > 0 {
		cfg.Token = u.User.String()
		u.User = nil
	}

	switch strings.ToLower(u.Scheme) {
	case "http":
		if cfg.Direct() {
			cfg.setDefaultPort(80)
		}
	case "https":
		if cfg.Direct() {
			cfg.setDefaultPort(443)
		}
	default:
		return nil, fmt.Errorf("invalid scheme '%s' in '%s', only 'http' is supported", u.Scheme, u)
	}

	if u.Path != "" {
		cfg.Path = u.Path
		u.Path = ""
	}

	if u.Port() != "" {
		port, err := strconv.Atoi(u.Port())
		if err != nil {
			return nil, fmt.Errorf("unable to convert port to integer: %s", err)
		}
		cfg.setDefaultPort(port)
		u.Host = u.Hostname()
	}
	cfg.Server = u.String()

	for _, set := range setters {
		set(cfg)
	}
	return cfg, nil
}

func newSender(cfg *configuration) (Sender, error) {
	client := internal.NewClient(cfg.Timeout, cfg.TLSConfig)
	metricsReporter := internal.NewReporter(cfg.metricsURL(), cfg.Token, client)
	tracesReporter := internal.NewReporter(cfg.tracesURL(), cfg.Token, client)
	sender := &wavefrontSender{
		defaultSource: internal.GetHostname("wavefront_direct_sender"),
		proxy:         !cfg.Direct(),
	}
	if cfg.InternalMetricsEnabled {
		sender.internalRegistry = sender.realInternalRegistry(cfg)
	} else {
		sender.internalRegistry = internal.NewNoOpRegistry()
	}
	sender.pointHandler = newLineHandler(metricsReporter, cfg, internal.MetricFormat, "points", sender.internalRegistry)
	sender.histoHandler = newLineHandler(metricsReporter, cfg, internal.HistogramFormat, "histograms", sender.internalRegistry)
	sender.spanHandler = newLineHandler(tracesReporter, cfg, internal.TraceFormat, "spans", sender.internalRegistry)
	sender.spanLogHandler = newLineHandler(tracesReporter, cfg, internal.SpanLogsFormat, "span_logs", sender.internalRegistry)
	sender.eventHandler = newLineHandler(metricsReporter, cfg, internal.EventFormat, "events", sender.internalRegistry)

	sender.Start()
	return sender, nil
}

func (c *configuration) tracesURL() string {
	return fmt.Sprintf("%s:%d%s", c.Server, c.TracesPort, c.Path)
}

func (c *configuration) metricsURL() string {
	return fmt.Sprintf("%s:%d%s", c.Server, c.MetricsPort, c.Path)
}

func (sender *wavefrontSender) realInternalRegistry(cfg *configuration) internal.MetricRegistry {
	var setters []internal.RegistryOption
	setters = append(setters, internal.SetInterval(cfg.InternalMetricsInterval))
	setters = append(setters, internal.SetPrefix(cfg.MetricPrefix()))
	setters = append(setters, internal.SetTag("pid", strconv.Itoa(os.Getpid())))
	setters = append(setters, internal.SetTag("version", version.Version))

	for key, value := range cfg.SDKMetricsTags {
		setters = append(setters, internal.SetTag(key, value))
	}

	return internal.NewMetricRegistry(
		sender,
		setters...,
	)

}

// BatchSize set max batch of data sent per flush interval. Defaults to 10,000. recommended not to exceed 40,000.
func BatchSize(n int) Option {
	return func(cfg *configuration) {
		cfg.BatchSize = n
	}
}

// MaxBufferSize set the size of internal buffers beyond which received data is dropped. Defaults to 50,000.
func MaxBufferSize(n int) Option {
	return func(cfg *configuration) {
		cfg.MaxBufferSize = n
	}
}

// FlushIntervalSeconds set the interval (in seconds) at which to flush data to Wavefront. Defaults to 1 Second.
func FlushIntervalSeconds(n int) Option {
	return func(cfg *configuration) {
		cfg.FlushInterval = time.Second * time.Duration(n)
	}
}

// FlushInterval set the interval at which to flush data to Wavefront. Defaults to 1 Second.
func FlushInterval(interval time.Duration) Option {
	return func(cfg *configuration) {
		cfg.FlushInterval = interval
	}
}

// MetricsPort sets the port on which to report metrics. Default is 2878.
func MetricsPort(port int) Option {
	return func(cfg *configuration) {
		cfg.MetricsPort = port
	}
}

// TracesPort sets the port on which to report traces. Default is 30001.
func TracesPort(port int) Option {
	return func(cfg *configuration) {
		cfg.TracesPort = port
	}
}

// Timeout sets the HTTP timeout (in seconds). Defaults to 10 seconds.
func Timeout(timeout time.Duration) Option {
	return func(cfg *configuration) {
		cfg.Timeout = timeout
	}
}

func TLSConfigOptions(tlsCfg *tls.Config) Option {
	tlsCfgCopy := tlsCfg.Clone()
	return func(cfg *configuration) {
		cfg.TLSConfig = tlsCfgCopy
	}
}

func InternalMetricsEnabled(enabled bool) Option {
	return func(cfg *configuration) {
		cfg.InternalMetricsEnabled = enabled
	}
}

func InternalMetricsInterval(interval time.Duration) Option {
	return func(cfg *configuration) {
		cfg.InternalMetricsInterval = interval
	}
}

// SDKMetricsTags adds the additional tags provided in tags to all internal
// metrics this library reports. Clients can use multiple SDKMetricsTags
// calls when creating a sender. In that case, the sender sends all the
// tags from each of the SDKMetricsTags calls in addition to the standard
// "pid" and "version" tags to all internal metrics. The "pid" tag is the
// process ID; the "version" tag is the version of this SDK.
func SDKMetricsTags(tags map[string]string) Option {
	// prevent caller from accidentally mutating this option.
	copiedTags := copyTags(tags)
	return func(cfg *configuration) {
		for key, value := range copiedTags {
			cfg.SDKMetricsTags[key] = value
		}
	}
}

func copyTags(orig map[string]string) map[string]string {
	result := make(map[string]string, len(orig))
	for key, value := range orig {
		result[key] = value
	}
	return result
}
