package senders_test

import (
	wavefront "github.com/wavefronthq/wavefront-sdk-go/senders"
)

func ExampleNewSender_proxy() {
	// For Proxy endpoints, by default metrics and histograms are sent to
	// port 2878 and spans are sent to port 30001.
	sender, err := wavefront.NewSender("http://localhost")
	if err != nil {
		// handle error
	}

	err = sender.Flush()
	if err != nil {
		// handle error
	}
	sender.Close()
}
