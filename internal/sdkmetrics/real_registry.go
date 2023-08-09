package sdkmetrics

import (
	"sync"
	"time"
)

// metric registry for internal metrics
type realMetricRegistry struct {
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

func (registry *realMetricRegistry) PointsTracker() SuccessTracker {
	return registry.pointsTracker
}

func (registry *realMetricRegistry) HistogramsTracker() SuccessTracker {
	return registry.histogramsTracker
}

func (registry *realMetricRegistry) SpansTracker() SuccessTracker {
	return registry.spansTracker
}

func (registry *realMetricRegistry) SpanLogsTracker() SuccessTracker {
	return registry.spanLogsTracker
}

func (registry *realMetricRegistry) EventsTracker() SuccessTracker {
	return registry.eventsTracker
}

type internalMetricFamily struct {
	Valid   *DeltaCounter
	Invalid *DeltaCounter
	Dropped *DeltaCounter
}

func (registry *realMetricRegistry) newInternalMetricFamily(prefix string) *internalMetricFamily {
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

func NewMetricRegistry(sender internalSender, setters ...RegistryOption) Registry {
	registry := &realMetricRegistry{
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

func (registry *realMetricRegistry) Start() {
	go registry.start()
}

func (registry *realMetricRegistry) start() {
	for {
		select {
		case <-registry.reportTicker.C:
			registry.report()
		case <-registry.done:
			return
		}
	}
}

func (registry *realMetricRegistry) Stop() {
	registry.reportTicker.Stop()
	registry.done <- struct{}{}
}

func (registry *realMetricRegistry) report() {
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

func (registry *realMetricRegistry) getOrAdd(name string, metric interface{}) interface{} {
	registry.mtx.Lock()
	defer registry.mtx.Unlock()

	if val, ok := registry.metrics[name]; ok {
		return val
	}
	registry.metrics[name] = metric
	return metric
}

func (registry *realMetricRegistry) NewCounter(name string) *MetricCounter {
	return registry.getOrAdd(name, &MetricCounter{}).(*MetricCounter)
}

func (registry *realMetricRegistry) NewDeltaCounter(name string) *DeltaCounter {
	return registry.getOrAdd(name, &DeltaCounter{MetricCounter{}}).(*DeltaCounter)
}

func (registry *realMetricRegistry) NewGauge(name string, f func() int64) *FunctionalGauge {
	return registry.getOrAdd(name, &FunctionalGauge{value: f}).(*FunctionalGauge)
}

func (registry *realMetricRegistry) NewGaugeFloat64(name string, f func() float64) *FunctionalGaugeFloat64 {
	return registry.getOrAdd(name, &FunctionalGaugeFloat64{value: f}).(*FunctionalGaugeFloat64)
}
