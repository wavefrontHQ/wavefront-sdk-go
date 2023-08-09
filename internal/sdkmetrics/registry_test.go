package sdkmetrics

import (
	"strings"
	"testing"
)

type fakeSender struct {
	count  int
	errors int
	prefix string
	name   string
	tags   []string
}

func (f *fakeSender) SendDeltaCounter(name string, value float64, source string, tags map[string]string) error {
	return nil
}

func (f *fakeSender) SendMetric(name string, value float64, ts int64, source string, tags map[string]string) error {
	f.count = f.count + 1
	if f.prefix != "" && !strings.HasPrefix(name, f.prefix) {
		f.errors = f.errors + 1
	}
	if f.name != "" && f.prefix != "" && (name != f.prefix+"."+f.name) {
		f.errors = f.errors + 1
	}
	for _, k := range f.tags {
		if tags == nil {
			f.errors = f.errors + 1
		} else {
			if _, ok := tags[k]; !ok {
				f.errors = f.errors + 1
			}
		}
	}
	return nil
}

func TestPrefix(t *testing.T) {
	prefix := "sdk.go.test"
	name := "test.counter"
	sender := &fakeSender{prefix: prefix, name: name}
	registry := NewMetricRegistry(sender, SetPrefix(prefix)).(*realMetricRegistry)
	registry.NewCounter(name)

	registry.report()
	if sender.errors != 0 {
		t.Error("name/prefix does not match")
	}
}

func TestRegistration(t *testing.T) {
	sender := &fakeSender{}
	registry := NewMetricRegistry(sender).(*realMetricRegistry)

	c := registry.NewCounter("counter")
	g := registry.NewGauge("gauge", func() int64 {
		return 100
	})

	if g.instantValue() != 100 {
		t.Error("unexpected gauge value")
	}

	c.Inc()
	if c.count() != 1 {
		t.Error("unexpected counter value")
	}

	registry.report()
	if sender.count != 2 {
		t.Error("unexpected number of metrics registered")
	}

	// verify same counter/gauges are returned
	altCounter := registry.NewCounter("counter")
	altCounter.Inc()
	if c.count() != 2 {
		t.Error("different counter returned")
	}

	altGauge := registry.NewGauge("gauge", func() int64 {
		return 1
	})
	if altGauge.instantValue() != 100 {
		t.Error("different gauge returned")
	}

	registry.report()
	if sender.count != 4 {
		t.Error("unexpected number of metrics registered")
	}
}

func TestTagging(t *testing.T) {
	sender := &fakeSender{tags: []string{"foo", "bar"}}
	registry := NewMetricRegistry(sender, SetTag("foo", "val"), SetTag("bar", "val")).(*realMetricRegistry)
	registry.NewCounter("counter")

	registry.report()
	if sender.errors != 0 {
		t.Error("tags do not match")
	}
}
