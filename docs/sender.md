# Set Up a Wavefront Sender

You can choose to send metrics, histograms, or trace data from your application to the Wavefront service using one of the following techniques:
* Use [direct ingestion](https://docs.wavefront.com/direct_ingestion.html) to send the data directly to the Wavefront service. This is the simplest way to get up and running quickly.
* Use a [Wavefront proxy](https://docs.wavefront.com/proxies.html), which then forwards the data to the Wavefront service. This is the recommended choice for a large-scale deployment that needs resilience to internet outages, control over data queuing and filtering, and more. 

To implement this choice: 

1. Import the `senders` package: 

    ```
    import (
        "github.com/wavefronthq/wavefront-sdk-go/senders"
    )
    ```
2. Create a `Sender` instance:
    * Option 1: [Create a direct `Sender`](#option-1-create-a-direct-sender) to send data directly to a Wavefront service.
    * Option 2: [Create a proxy `Sender`](#option-2-create-a-proxy-sender) to send data to a Wavefront proxy.

## Option 1: Create a Direct Sender
To create a direct `Sender`, you configure it to send data directly to Wavefront.

### Step 1. Obtain Wavefront Access Information
Gather the following access information:

* Identify the URL of your Wavefront instance. This is the URL you connect to when you log in to Wavefront, typically something like `https://<domain>.wavefront.com`.
* In Wavefront, verify that you have Direct Data Ingestion permission, and [obtain an API token](http://docs.wavefront.com/wavefront_api.html#generating-an-api-token).

### Step 2. Configure the Direct Sender
Create a `DirectConfiguration` that has the server and token information you obtained in Step 1. 

You can optionally tune the following ingestion properties:

* Max buffer size - Internal buffer capacity of the direct `Sender`, which helps with handling brief increases in data and buffering on errors. Any data in excess of this size is dropped.  Separate buffers are maintained per data type (metrics, spans and distributions).
* Flush interval - Interval for flushing data from the direct `Sender` directly to Wavefront.
* Batch size - Amount of data to send to Wavefront in each flush interval.

Together, the flush interval and batch size control the maximum theoretical throughput of the direct `Sender`. 

Default values should suffice for most use cases.
You should override the defaults _only_ to set higher values.

```go
import (
    time
    "github.com/wavefronthq/wavefront-sdk-go/senders"
)

func main() {
    directCfg := &senders.DirectConfiguration {
        // Your Wavefront instance URL
        Server : "https://INSTANCE.wavefront.com", 
        
        // Wavefront API token created with direct ingestion permission
        Token : "YOUR_API_TOKEN",

        // Optional: Override the batch size (in data points). Default: 10,000. Recommended not to exceed 40,000.
        BatchSize : 20000,

        // Optional: Override the max buffer size (in data points). Default: 50,000. Higher values could use more memory.
        MaxBufferSize : 50000,

        // Optional: Override the flush interval (in seconds). Default: 1 second
        FlushIntervalSeconds : 2,
    }

    // Create the direct sender
    sender, err := senders.NewDirectSender(directCfg)
    if err != nil {
        // handle error
    }
    ... // Use the direct sender to send data 
}
```


## Option 2: Create a Proxy Sender

**Note:** Before your application can use a proxy `Sender`, you must [set up and start a Wavefront proxy](https://github.com/wavefrontHQ/java/tree/master/proxy#set-up-a-wavefront-proxy).

To create a proxy `Sender`, you configure it with a `ProxyConfiguration` that includes:

* The name of the host that will run the Wavefront proxy.
* One or more proxy listening ports to send data to. The ports you specify depend on the kinds of data you want to send (metrics, histograms, and/or trace data). You must specify at least one listener port. 
* Optional setting for tuning communication with the proxy.

```go
import (
    "github.com/wavefronthq/wavefront-sdk-go/senders"
)

func main() {
    proxyCfg := &senders.ProxyConfiguration {
        // The proxy hostname or address
        Host : "proxyHostname or proxyIPAddress",

        // Set at least one port:
        
        // Set the proxy port to send metrics to. Default: 2878
        MetricsPort : 2878, 
        
        // Set a proxy port to send histograms to.  Recommended: 2878
        DistributionPort: 2878,
        
        // Set a proxy port to send trace data to. Recommended: 30000
        TracingPort : 30000,
        
        // Optional: Set a nondefault interval (in seconds) for flushing data from the sender to the proxy. Default: 5 seconds
        FlushIntervalSeconds: 10 
        
    }

    // Create the proxy sender
    sender, err := senders.NewProxySender(proxyCfg)
    if err != nil {
        // handle error
    }
    
    ... // Use the proxy sender to send data 
}
```

**Note:** When you [set up a Wavefront proxy](https://github.com/wavefrontHQ/java/tree/master/proxy#set-up-a-wavefront-proxy) on the specified proxy host, you specify the port it will listen to for each type of data to be sent. The proxy `Sender` must send data to the same ports that the Wavefront proxy listens to. Consequently, the port-related `Sender` configuration fields must specify the same port numbers as the corresponding proxy configuration properties: 

| `ProxyConfiguration` field | Corresponding property in `wavefront.conf` |
| ----- | -------- |
| `MetricsPort` | `pushListenerPorts=` |
| `DistributionPort` | `histogramDistListenerPorts=` |
| `TracingPort` | `traceListenerPorts=` |
 
# Share a Wavefront Sender

Various Wavefront SDKs for Go use this library and require a `Sender` instance.

If you are using multiple Wavefront Go SDKs within the same process, you can create a single `Sender` and share it among the SDKs. 
 
<!--- 
For example, the following snippet shows how to use the same `Sender` when setting up the [wavefront-opentracing-sdk-go](https://github.com/wavefrontHQ/wavefront-opentracing-sdk-go) and XXX SDKs.

```
```
--->
**Note:** If you use SDKs in different processes, you must create one `Sender` instance per process.
