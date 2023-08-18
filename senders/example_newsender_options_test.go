package senders_test

import (
	"crypto/tls"
	wavefront "github.com/wavefronthq/wavefront-sdk-go/senders"
	"time"
)

func ExampleNewSender_options() {
	// NewSender accepts optional arguments. Use these if you need to set non-default ports for your Wavefront Proxy, tune batching parameters, or set tags for internal SDK metrics.
	sender, err := wavefront.NewSender(
		"http://localhost",
		wavefront.BatchSize(20000),                // Send batches of 20,000.
		wavefront.FlushInterval(5*time.Second),    // Flush every 5 seconds.
		wavefront.MetricsPort(4321),               // Use port 4321 for metrics.
		wavefront.TracesPort(40001),               // Use port 40001 for traces.
		wavefront.Timeout(15),                     // Set an HTTP timeout in seconds (default is 10s)
		wavefront.SendInternalMetrics(false),      // Don't send internal ~sdk.go.* metrics
		wavefront.TLSConfigOptions(&tls.Config{}), // Set TLS config options.
	)

	if err != nil {
		// handle error
	}
	sender.Close()
}
