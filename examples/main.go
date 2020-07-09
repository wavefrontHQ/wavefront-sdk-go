package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/wavefronthq/wavefront-sdk-go/application"
	"github.com/wavefronthq/wavefront-sdk-go/event"

	"github.com/wavefronthq/wavefront-sdk-go/senders"
)

func main() {
	wf := senders.NewMultiSender()

	urls := strings.Split(os.Getenv("WF_URL"), "|")
	for idx, url := range urls {
		sender, err := senders.NewSender(url)
		if err != nil {
			panic(err)
		}
		sender.SetSource(fmt.Sprintf("sender_%d", idx))
		wf.Add(sender)
	}

	// OLD PROXY way
	proxyCfg := &senders.ProxyConfiguration{
		Host:                 "localhost",
		MetricsPort:          2878,
		DistributionPort:     2878,
		TracingPort:          2878,
		EventsPort:           2878,
		FlushIntervalSeconds: 10,
	}

	sender, err := senders.NewProxySender(proxyCfg)
	if err != nil {
		panic(err)
	}
	sender.SetSource("client_proxy")
	wf.Add(sender)

	// OLD DIRECT way
	directCfg := &senders.DirectConfiguration{
		Server:               "https://-----.wavefront.com",
		Token:                "--------------",
		BatchSize:            10000,
		MaxBufferSize:        50000,
		FlushIntervalSeconds: 1,
	}

	sender, err = senders.NewDirectSender(directCfg)
	if err != nil {
		panic(err)
	}
	sender.SetSource("client_direct")
	wf.Add(sender)

	log.Print("senders ready")

	source := "" //"go_sdk_example"

	app := application.New("sample app", "main.go")
	application.StartHeartbeatService(wf, app, source)

	tags := make(map[string]string)
	tags["namespace"] = "default"
	tags["Kind"] = "Deployment"

	options := []event.Option{event.Details("Details"), event.Type("type"), event.Severity("severity")}

	for i := 0; i < 10; i++ {
		wf.SendMetric("sample.metric", float64(i), time.Now().UnixNano(), source, map[string]string{"env": "test"})

		txt := fmt.Sprintf("test event %d", i)
		sendEvent(wf, txt, time.Now().Unix(), 0, source, tags, options...)

		time.Sleep(10 * time.Second)
	}

	wf.Flush()
	wf.Close()
}

func sendEvent(sender senders.Sender, name string, startMillis, endMillis int64, source string, tags map[string]string, setters ...event.Option) {
	err := sender.SendEvent(name, startMillis, endMillis, source, tags, setters...)
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
}
