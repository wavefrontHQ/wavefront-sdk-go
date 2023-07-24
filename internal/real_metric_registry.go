package internal

import (
	"sync"
	"time"
)

// metric registry for internal metrics
type RealMetricRegistry struct {
	source       string
	prefix       string
	tags         map[string]string
	reportTicker *time.Ticker
	sender       internalSender
	done         chan struct{}

	mtx               sync.Mutex
	metrics           map[string]interface{}
	pointsTracker     *internalMetricFamily
	histogramsTracker *internalMetricFamily
	spansTracker      *internalMetricFamily
	spanLogsTracker   *internalMetricFamily
	eventsTracker     *internalMetricFamily
}

func (registry *RealMetricRegistry) PointsTracker() SuccessTracker {
	return registry.pointsTracker
}

func (registry *RealMetricRegistry) HistogramsTracker() SuccessTracker {
	return registry.histogramsTracker
}

func (registry *RealMetricRegistry) SpansTracker() SuccessTracker {
	return registry.spansTracker
}

func (registry *RealMetricRegistry) SpanLogsTracker() SuccessTracker {
	return registry.spanLogsTracker
}

func (registry *RealMetricRegistry) EventsTracker() SuccessTracker {
	return registry.eventsTracker
}

type internalMetricFamily struct {
	Valid   *DeltaCounter
	Invalid *DeltaCounter
	Dropped *DeltaCounter
}

func (registry *RealMetricRegistry) newInternalMetricFamily(prefix string) *internalMetricFamily {
	return &internalMetricFamily{
		Valid:   registry.NewDeltaCounter(prefix + ".valid"),
		Invalid: registry.NewDeltaCounter(prefix + ".invalid"),
		Dropped: registry.NewDeltaCounter(prefix + ".dropped"),
	}
}

type SuccessTracker interface {
	IncValid()
	IncInvalid()
	IncDropped()
}

func (f *internalMetricFamily) IncValid() {
	f.Valid.Inc()
}

func (f *internalMetricFamily) IncInvalid() {
	f.Invalid.Inc()
}

func (f *internalMetricFamily) IncDropped() {
	f.Dropped.Inc()
}

func NewMetricRegistry(sender internalSender, setters ...RegistryOption) *RealMetricRegistry {
	registry := &RealMetricRegistry{
		sender:       sender,
		metrics:      make(map[string]interface{}),
		reportTicker: time.NewTicker(time.Second * 60),
		done:         make(chan struct{}),
	}

	registry.pointsTracker = registry.newInternalMetricFamily("points")
	registry.histogramsTracker = registry.newInternalMetricFamily("histograms")
	registry.spansTracker = registry.newInternalMetricFamily("spans")
	registry.spanLogsTracker = registry.newInternalMetricFamily("span_logs")
	registry.eventsTracker = registry.newInternalMetricFamily("events")

	for _, setter := range setters {
		setter(registry)
	}
	return registry
}

func (registry *RealMetricRegistry) Start() {
	go registry.start()
}

func (registry *RealMetricRegistry) start() {
	for {
		select {
		case <-registry.reportTicker.C:
			registry.report()
		case <-registry.done:
			return
		}
	}
}

func (registry *RealMetricRegistry) Stop() {
	registry.reportTicker.Stop()
	registry.done <- struct{}{}
}

func (registry *RealMetricRegistry) report() {
	registry.mtx.Lock()
	defer registry.mtx.Unlock()

	for k, metric := range registry.metrics {
		switch metric.(type) {
		case *DeltaCounter:
			deltaCount := metric.(*DeltaCounter).count()
			registry.sender.SendDeltaCounter(registry.prefix+"."+k, float64(deltaCount), "", registry.tags)
			metric.(*DeltaCounter).dec(deltaCount)
		case *MetricCounter:
			registry.sender.SendMetric(registry.prefix+"."+k, float64(metric.(*MetricCounter).count()), 0, "", registry.tags)
		case *FunctionalGauge:
			registry.sender.SendMetric(registry.prefix+"."+k, float64(metric.(*FunctionalGauge).instantValue()), 0, "", registry.tags)
		case *FunctionalGaugeFloat64:
			registry.sender.SendMetric(registry.prefix+"."+k, metric.(*FunctionalGaugeFloat64).instantValue(), 0, "", registry.tags)
		}
	}
}

func (registry *RealMetricRegistry) getOrAdd(name string, metric interface{}) interface{} {
	registry.mtx.Lock()
	defer registry.mtx.Unlock()

	if val, ok := registry.metrics[name]; ok {
		return val
	}
	registry.metrics[name] = metric
	return metric
}

func (registry *RealMetricRegistry) NewCounter(name string) *MetricCounter {
	return registry.getOrAdd(name, &MetricCounter{}).(*MetricCounter)
}

func (registry *RealMetricRegistry) NewDeltaCounter(name string) *DeltaCounter {
	return registry.getOrAdd(name, &DeltaCounter{MetricCounter{}}).(*DeltaCounter)
}

func (registry *RealMetricRegistry) NewGauge(name string, f func() int64) *FunctionalGauge {
	return registry.getOrAdd(name, &FunctionalGauge{value: f}).(*FunctionalGauge)
}

func (registry *RealMetricRegistry) NewGaugeFloat64(name string, f func() float64) *FunctionalGaugeFloat64 {
	return registry.getOrAdd(name, &FunctionalGaugeFloat64{value: f}).(*FunctionalGaugeFloat64)
}

type RegistryOption func(*RealMetricRegistry)

func SetSource(source string) RegistryOption {
	return func(registry *RealMetricRegistry) {
		registry.source = source
	}
}

func SetInterval(interval time.Duration) RegistryOption {
	return func(registry *RealMetricRegistry) {
		registry.reportTicker = time.NewTicker(interval)
	}
}

func SetTags(tags map[string]string) RegistryOption {
	return func(registry *RealMetricRegistry) {
		registry.tags = tags
	}
}

func SetTag(key, value string) RegistryOption {
	return func(registry *RealMetricRegistry) {
		if registry.tags == nil {
			registry.tags = make(map[string]string)
		}
		registry.tags[key] = value
	}
}

func SetPrefix(prefix string) RegistryOption {
	return func(registry *RealMetricRegistry) {
		registry.prefix = prefix
	}
}
