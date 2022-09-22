package senders_test

import (
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
	wavefront "github.com/wavefronthq/wavefront-sdk-go/senders"
)

var sender wavefront.Sender

func ExampleMetricSender_SendMetric() {
	// Wavefront metrics data format
	// <metricName> <metricValue> [<timestamp>] source=<source> [pointTags]
	// Example: "new-york.power.usage 42422 1533529977 source=localhost datacenter=dc1"
	err := sender.SendMetric("new-york.power.usage", 42422.0, 0, "go_test", map[string]string{"env": "test"})
	if err != nil {
		// handle err
	}
}

func ExampleSpanSender_SendSpan() {
	// When you use a Sender SDK, you won’t see span-level RED metrics by default unless you use the Wavefront proxy and define a custom tracing port (TracingPort). See Instrument Your Application with Wavefront Sender SDKs for details.
	// Wavefront Tracing Span Data format
	// <tracingSpanName> source=<source> [pointTags] <start_millis> <duration_milliseconds>
	// Example:
	// "getAllUsers source=localhost traceId=7b3bf470-9456-11e8-9eb6-529269fb1459
	// spanId=0313bafe-9457-11e8-9eb6-529269fb1459 parent=2f64e538-9457-11e8-9eb6-529269fb1459
	// application=Wavefront http.method=GET 1552949776000 343"
	err := sender.SendSpan("getAllUsers", 1552949776000, 343, "localhost",
		"7b3bf470-9456-11e8-9eb6-529269fb1459",
		"0313bafe-9457-11e8-9eb6-529269fb1459",
		[]string{"2f64e538-9457-11e8-9eb6-529269fb1459"},
		nil,
		[]wavefront.SpanTag{
			{Key: "application", Value: "Wavefront"},
			{Key: "service", Value: "istio"},
			{Key: "http.method", Value: "GET"},
		},
		nil)
	if err != nil {
		// handle err
	}
}

func ExampleMetricSender_SendDeltaCounter() {
	// Wavefront delta counter format
	// <metricName> <metricValue> source=<source> [pointTags]
	// Example: "lambda.thumbnail.generate 10 source=thumbnail_service image-format=jpeg"
	err := sender.SendDeltaCounter("lambda.thumbnail.generate", 10.0, "thumbnail_service", map[string]string{"format": "jpeg"})
	if err != nil {
		// handle err
	}
}

func ExampleDistributionSender_SendDistribution() {
	// Wavefront Histogram data format
	// {!M | !H | !D} [<timestamp>] #<count> <mean> [centroids] <histogramName> source=<source> [pointTags]
	// Example: You can choose to send to at most 3 bins - Minute/Hour/Day
	// "!M 1533529977 #20 30.0 #10 5.1 request.latency source=appServer1 region=us-west"
	// "!H 1533529977 #20 30.0 #10 5.1 request.latency source=appServer1 region=us-west"
	// "!D 1533529977 #20 30.0 #10 5.1 request.latency source=appServer1 region=us-west"

	centroids := []histogram.Centroid{
		{
			Value: 30.0,
			Count: 20,
		},
		{
			Value: 5.1,
			Count: 10,
		},
	}

	hgs := map[histogram.Granularity]bool{
		histogram.MINUTE: true,
		histogram.HOUR:   true,
		histogram.DAY:    true,
	}

	err := sender.SendDistribution("request.latency", centroids, hgs, 0, "appServer1", map[string]string{"region": "us-west"})
	if err != nil {
		// handle err
	}
}
