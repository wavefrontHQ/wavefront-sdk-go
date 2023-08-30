package senders_test

import (
	wavefront "github.com/wavefronthq/wavefront-sdk-go/senders"
)

func ExampleNewSender_proxy() {
	// For Proxy endpoints, by default metrics and histograms are sent to
	// port 2878 and traces are sent to port 30001.
	sender, err := wavefront.NewSender("http://localhost")

	// To use non-default ports, specify the metric/histogram port in the url,
	// and set the port for traces using the TracesPort Option.
	sender, err = wavefront.NewSender("https://localhost:4443",
		wavefront.TracesPort(55555))

	if err != nil {
		// handle error
	}

	err = sender.Flush()
	if err != nil {
		// handle error
	}
	sender.Close()
}
