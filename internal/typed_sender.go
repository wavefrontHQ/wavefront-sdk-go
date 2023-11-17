package internal

import "github.com/wavefronthq/wavefront-sdk-go/internal/sdkmetrics"

type TypedSender interface {
	TrySend(string, error) error
	Start()
	Stop()
	Flush() error
	GetFailureCount() int64
}

type typedSender struct {
	tracker     sdkmetrics.SuccessTracker
	lineHandler BatchBuilder
}

func (ts *typedSender) Start() {
	ts.lineHandler.Start()
}

func (ts *typedSender) Stop() {
	ts.lineHandler.Stop()
}

func (ts *typedSender) Flush() error {
	return ts.lineHandler.Flush()
}

func (ts *typedSender) GetFailureCount() int64 {
	return ts.lineHandler.GetFailureCount()
}

func (ts *typedSender) TrySend(line string, err error) error {
	if err != nil {
		ts.tracker.IncInvalid()
		return err
	}

	ts.tracker.IncValid()
	err = ts.lineHandler.HandleLine(line)
	if err != nil {
		ts.tracker.IncDropped()
	}
	return err
}

func NewTypedSender(tracker sdkmetrics.SuccessTracker, handler BatchBuilder) TypedSender {
	return &typedSender{
		tracker:     tracker,
		lineHandler: handler,
	}
}
