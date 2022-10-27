package senders_test

import wavefront "github.com/wavefronthq/wavefront-sdk-go/senders"

func ExampleNewSender_direct() {
	// For Direct Ingestion endpoints, by default all data is sent to port 80
	// or port 443 for unencrypted or encrypted connections, respectively.
	// 11111111-2222-3333-4444-555555555555 is your API token with direct ingestion permission.
	sender, err := wavefront.NewSender("https://11111111-2222-3333-4444-555555555555@surf.wavefront.com")
	if err != nil {
		// handle error
	}
	err = sender.Flush()
	if err != nil {
		// handle error
	}
	sender.Close()
}
