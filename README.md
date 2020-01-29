# wavefront-sdk-go [![build status][ci-img]][ci] [![Go Report Card][go-report-img]][go-report] [![GoDoc][godoc-img]][godoc]

This library provides support for sending metrics, histograms and tracing spans to Wavefront via proxy or direct ingestion using the `Sender` interface.

## Requirements
- Go 1.9 or higher

## Usage

Import the `senders` package and create a proxy or direct sender as given below.

```go
import (
    wavefront "github.com/wavefronthq/wavefront-sdk-go/senders"
)
```

Make sure you have go.mod file that can be updated. If you do not have the go.mod file, run the following command to initialize it.

```
go mod init
```

Sometimes, it is necessary for you to provide the module name when you are initializing it (usually the directory that you are running go application).

```
go mod init <module_name>
```

### Proxy Sender
Depending on the data you wish to send to Wavefront (metrics, distributions and/or spans), enable the relevant ports on the proxy and initialize the proxy sender as follows:

```go
import (
    wavefront "github.com/wavefronthq/wavefront-sdk-go/senders"
)

func main() {
    proxyCfg := &wavefront.ProxyConfiguration {
        Host : "proxyHostname or proxyIPAddress",

        // At least one port should be set below.
        MetricsPort : 2878,      // set this (typically 2878) to send metrics
        DistributionPort: 2878,  // set this (typically 2878) to send distributions
        TracingPort : 30000,     // set this to send tracing spans

        FlushIntervalSeconds: 10, // flush the buffer periodically, defaults to 5 seconds.
    }

    sender, err := wavefront.NewProxySender(proxyCfg)
    if err != nil {
        // handle error
    }
    // send data (see below for usage)
    
    time.Sleep(5 * time.Second)
	sender.Flush()
	sender.Close()
}
```

### Direct Sender

```go
import (
    wavefront "github.com/wavefronthq/wavefront-sdk-go/senders"
)

func main() {
    directCfg := &wavefront.DirectConfiguration {
        Server : "https://INSTANCE.wavefront.com", // your Wavefront instance URL
        Token : "YOUR_API_TOKEN",                  // API token with direct ingestion permission

        // Optional configuration properties. Default values should suffice for most use cases.
        // override the defaults only if you wish to set higher values.

        // max batch of data sent per flush interval. defaults to 10,000.
        // recommended not to exceed 40,000.
        BatchSize : 10000,

        // size of internal buffer beyond which received data is dropped.
        // helps with handling brief increases in data and buffering on errors.
        // separate buffers are maintained per data type (metrics, spans and distributions)
        // defaults to 50,000. higher values could use more memory.
        MaxBufferSize : 50000,

        // interval (in seconds) at which to flush data to Wavefront. defaults to 1 Second.
        // together with batch size controls the max theoretical throughput of the sender.
        FlushIntervalSeconds : 1,
    }

    sender, err := wavefront.NewDirectSender(directCfg)
    if err != nil {
        // handle error
    }
    // send data (see below for usage)
    
    time.Sleep(5 * time.Second)
    sender.Flush()
    sender.Close()
}

```

### Sending data to Wavefront

Use the `Sender` interface for sending data to Wavefront.

#### Metrics and Delta Counters

```go
// Wavefront metrics data format
// <metricName> <metricValue> [<timestamp>] source=<source> [pointTags]
// Example: "new-york.power.usage 42422 1533529977 source=localhost datacenter=dc1"
sender.SendMetric("new-york.power.usage", 42422.0, 0, "go_test", map[string]string{"env" : "test"})

// Wavefront delta counter format
// <metricName> <metricValue> source=<source> [pointTags]
// Example: "lambda.thumbnail.generate 10 source=thumbnail_service image-format=jpeg"
sender.SendDeltaCounter("lambda.thumbnail.generate", 10.0, "thumbnail_service", map[string]string{"format" : "jpeg"})
```

#### Distributions

```go
import "github.com/wavefronthq/wavefront-sdk-go/histogram"

// Wavefront Histogram data format
// {!M | !H | !D} [<timestamp>] #<count> <mean> [centroids] <histogramName> source=<source> [pointTags]
// Example: You can choose to send to at most 3 bins - Minute/Hour/Day
// "!M 1533529977 #20 30.0 #10 5.1 request.latency source=appServer1 region=us-west"
// "!H 1533529977 #20 30.0 #10 5.1 request.latency source=appServer1 region=us-west"
// "!D 1533529977 #20 30.0 #10 5.1 request.latency source=appServer1 region=us-west"

centroids := []histogram.Centroid {
      {
        Value : 30.0,
        Count : 20,
      },
      {
        Value : 5.1,
        Count : 10,
      },
}

hgs := map[histogram.Granularity]bool {
    histogram.MINUTE : true,
    histogram.HOUR   : true,
    histogram.DAY    : true,
}

sender.SendDistribution("request.latency", centroids, hgs, 0, "appServer1", map[string]string {"region" : "us-west"})
```

#### Tracing Spans

```go
// Wavefront Tracing Span Data format
// <tracingSpanName> source=<source> [pointTags] <start_millis> <duration_milliseconds>
// Example:
// "getAllUsers source=localhost traceId=7b3bf470-9456-11e8-9eb6-529269fb1459
// spanId=0313bafe-9457-11e8-9eb6-529269fb1459 parent=2f64e538-9457-11e8-9eb6-529269fb1459
// application=Wavefront http.method=GET 1552949776000 343"

sender.SendSpan("getAllUsers", 1552949776000, 343, "localhost",
    "7b3bf470-9456-11e8-9eb6-529269fb1459",
    "0313bafe-9457-11e8-9eb6-529269fb1459",
    []string {"2f64e538-9457-11e8-9eb6-529269fb1459"},
    nil,
    []SpanTag {
        {Key : "application", Value : "Wavefront"},
        {Key : "http.method", Value : "GET"},
    },
    nil)
```
**Note:** The tracing and span SDK APIs are designed to serve as low-level endpoints. For most use cases, we recommend using
opentracing with the ```WavefrontTracer```.

For more information on OpenTracing, please refer the OpenTracing project: https://github.com/opentracing/opentracing-go. To use OpenTracing with Wavefront, please refer to https://github.com/wavefrontHQ/wavefront-opentracing-sdk-go.

#### Closing the Sender
It is recommended to flush and close the sender before shutting down your application.

```go
// failures observed while sending metrics/histograms/spans, can be obtained as follows:
totalFailures := sender.GetFailureCount()

// on-demand buffer flush
sender.Flush()

// close the sender before shutting down your application
sender.Close()
```

[ci-img]: https://travis-ci.com/wavefrontHQ/wavefront-sdk-go.svg?branch=master
[ci]: https://travis-ci.com/wavefrontHQ/wavefront-sdk-go
[godoc]: https://godoc.org/github.com/wavefrontHQ/wavefront-sdk-go/senders
[godoc-img]: https://godoc.org/github.com/wavefrontHQ/wavefront-sdk-go/senders?status.svg
[go-report-img]: https://goreportcard.com/badge/github.com/wavefronthq/wavefront-sdk-go
[go-report]: https://goreportcard.com/report/github.com/wavefronthq/wavefront-sdk-go
