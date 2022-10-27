package senders

import (
	"github.com/wavefronthq/wavefront-sdk-go/event"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
)

// MetricSender Interface for sending metrics to Wavefront
type MetricSender interface {
	// Sends a single metric to Wavefront with optional timestamp and tags.
	SendMetric(name string, value float64, ts int64, source string, tags map[string]string) error

	// Sends a delta counter (counter aggregated at the Wavefront service) to Wavefront.
	// the timestamp for a delta counter is assigned at the server side.
	SendDeltaCounter(name string, value float64, source string, tags map[string]string) error
}

// DistributionSender Interface for sending distributions to Wavefront
type DistributionSender interface {
	// Sends a distribution of metrics to Wavefront with optional timestamp and tags.
	// Each centroid is a 2-dimensional entity with the first dimension the mean value
	// and the second dimension the count of points in the centroid.
	// The granularity informs the set of intervals (minute, hour, and/or day) by which the
	// histogram data should be aggregated.
	SendDistribution(name string, centroids []histogram.Centroid, hgs map[histogram.Granularity]bool, ts int64, source string, tags map[string]string) error
}

// EventSender Interface for sending events to Wavefront. NOT yet supported.
type EventSender interface {
	// Sends an event to Wavefront with optional tags
	SendEvent(name string, startMillis, endMillis int64, source string, tags map[string]string, setters ...event.Option) error
}
