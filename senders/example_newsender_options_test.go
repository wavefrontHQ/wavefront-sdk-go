package senders_test

import (
	wavefront "github.com/wavefronthq/wavefront-sdk-go/senders"
)

func ExampleNewSender_options() {
	// NewSender accepts optional arguments. Use these if you need to set non-default ports for your Wavefront Proxy, tune batching parameters, or set tags for internal SDK metrics.
	sender, err := wavefront.NewSender(
		"http://localhost",
		wavefront.BatchSize(20000),        // Send batches of 20,000.
		wavefront.FlushIntervalSeconds(5), // Flush every 5 seconds.
		wavefront.MetricsPort(4321),       // Use port 4321 for metrics.
		wavefront.TracesPort(40001),       // Use port 40001 for traces.
	)

	if err != nil {
		// handle error
	}
	sender.Close()
}
