package senders_test

import wavefront "github.com/wavefronthq/wavefront-sdk-go/senders"

func ExampleNewSender_direct() {
	// For Direct Ingestion endpoints, by default all data is sent to port 80
	// or port 443 for unencrypted or encrypted connections, respectively.

	// Direct Ingestion requires authentication.

	// Wavefront API tokens:
	// Set your API token using the APIToken Option
	// 11111111-2222-3333-4444-555555555555 is your API token with direct ingestion permission.
	sender, err := wavefront.NewSender("https://surf.wavefront.com",
		wavefront.APIToken("11111111-2222-3333-4444-555555555555"))

	// CSP API tokens:
	// Set your API token using the CSPAPIToken Option
	// <MY-CSP-TOKEN> is your CSP API token with the aoa:directDataIngestion scope.
	sender, err = wavefront.NewSender("https://surf.wavefront.com",
		wavefront.CSPAPIToken("<MY-CSP-TOKEN>"))

	// CSP Client Credentials:
	// Set your API token using the CSPClientCredentials Option
	sender, err = wavefront.NewSender("https://surf.wavefront.com",
		wavefront.CSPClientCredentials("<MY-CLIENT_ID>", "<MY-CLIENT_SECRET>"))

	// CSP Authentication strategies also support additional options
	sender, err = wavefront.NewSender("https://surf.wavefront.com",
		wavefront.CSPClientCredentials(
			"<MY-CLIENT_ID>",
			"<MY-CLIENT_SECRET>",
			wavefront.CSPBaseURL("<MY-NONSTANDARD-CSP-HOST>"),
			wavefront.CSPOrgID("<MY-ORG-ID>"),
		))

	sender, err = wavefront.NewSender("https://surf.wavefront.com",
		wavefront.CSPAPIToken(
			"<MY-CSP-TOKEN>",
			wavefront.CSPBaseURL("<MY-NONSTANDARD-CSP-HOST>"),
		))

	if err != nil {
		// handle error
	}
	err = sender.Flush()
	if err != nil {
		// handle error
	}
	sender.Close()
}
