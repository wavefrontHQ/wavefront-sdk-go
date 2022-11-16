# wavefront-sdk-go

[![CI Status](https://github.com/wavefrontHQ/wavefront-sdk-go/actions/workflows/main.yml/badge.svg)](https://github.com/wavefrontHQ/wavefront-sdk-go/actions/workflows/main.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/wavefronthq/wavefront-sdk-go.svg)](https://pkg.go.dev/github.com/wavefronthq/wavefront-sdk-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/wavefronthq/wavefront-sdk-go)](https://goreportcard.com/report/github.com/wavefronthq/wavefront-sdk-go)

## Table of Contents
* [Internal SDK Metrics](#internal-sdk-metrics)
* [License](#License)
* [How to Get Support and Contribute](#how-to-get-support-and-contribute)

# Welcome to the Wavefront Go SDK

Wavefront by VMware Go SDK lets you send raw data from your Go application to Wavefront using a `Sender` interface.
The data is then stored as metrics, histograms, and trace data. This SDK is also called the Wavefront Sender SDK for Go. 

Although this library is mostly used by the other Wavefront Go SDKs to send data to Wavefront, 
you can also use this SDK directly. For example, you can send data directly from a data store or CSV file to Wavefront.

To learn more about how to send data, the SDK types, and functions, see [pkg.go.dev documentation](https://pkg.go.dev/github.com/wavefronthq/wavefront-sdk-go)

# Internal SDK Metrics

The SDK optionally adds its own metrics. The internal metrics are prefixed with `~sdk.go.core.sender.direct` or  `~sdk.go.core.sender.proxy`, depending on whether metrics are being sent directly or via a Wavefront Proxy.

| metric name          |
|----------------------|
| `points.valid`       |
| `points.invalid`     |  
| `points.dropped`     |  
| `histograms.valid`   | 
| `histograms.invalid` |
| `histograms.dropped` |
| `spans.valid`        |
| `spans.invalid`      |
| `spans.dropped`      |
| `span_logs.valid`    |
| `span_logs.invalid`  |
| `span_logs.dropped`  |
| `events.valid`       |
| `events.invalid`     |
| `events.dropped`     |

## License
[Apache 2.0 License](LICENSE).

## How to Get Support and Contribute

* Reach out to us on our public [Slack channel](https://www.wavefront.com/join-public-slack).
* If you run into any issues, let us know by creating a GitHub issue.

To create a new release, follow the instructions in [RELEASING.md](RELEASING.md)