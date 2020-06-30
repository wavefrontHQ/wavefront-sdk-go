package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/wavefronthq/wavefront-sdk-go/application"
	"github.com/wavefronthq/wavefront-sdk-go/event"
	"github.com/wavefronthq/wavefront-sdk-go/wavefront"
)

func main() {
	wf, err := wavefront.NewClient(os.Getenv("WF_URL"))
	if err != nil {
		panic(err)
	}
	log.Print("sender ready")

	source := "go_sdk_example"

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

func sendEvent(sender wavefront.Client, name string, startMillis, endMillis int64, source string, tags map[string]string, setters ...event.Option) {
	err := sender.SendEvent(name, startMillis, endMillis, source, tags, setters...)
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
}
