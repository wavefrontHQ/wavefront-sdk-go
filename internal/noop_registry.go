package internal

type NoOpRegistry struct{}

type noOpTracker struct {
}

func (n noOpTracker) IncValid() {
}

func (n noOpTracker) IncInvalid() {
}

func (n noOpTracker) IncDropped() {
}

func (n *NoOpRegistry) PointsTracker() SuccessTracker {
	return &noOpTracker{}
}

func (n *NoOpRegistry) HistogramsTracker() SuccessTracker {
	return &noOpTracker{}

}

func (n *NoOpRegistry) SpansTracker() SuccessTracker {
	return &noOpTracker{}
}

func (n *NoOpRegistry) SpanLogsTracker() SuccessTracker {
	return &noOpTracker{}
}

func (n *NoOpRegistry) EventsTracker() SuccessTracker {
	return &noOpTracker{}
}

func NewNoOpRegistry() MetricRegistry {
	return &NoOpRegistry{}
}

func (n *NoOpRegistry) Start() {
}

func (n *NoOpRegistry) Stop() {
}

func (n *NoOpRegistry) NewGauge(s string, f func() int64) *FunctionalGauge {
	return &FunctionalGauge{}
}
