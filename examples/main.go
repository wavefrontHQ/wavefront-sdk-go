package main

import (
	"log"
	"os"
	"time"

	"github.com/wavefronthq/wavefront-sdk-go/event"
	"github.com/wavefronthq/wavefront-sdk-go/wavefront"
)

func main() {
	wf, err := wavefront.NewClient(os.Getenv("WF_URL"))
	if err != nil {
		panic(err)
	}

	source := "gosdk_histogram_test"

	log.Print("sender ready")

	tags := make(map[string]string)
	tags["namespace"] = "default"
	tags["Kind"] = "Deployment"

	options := []event.Option{event.Details("Details"), event.Type("type"), event.Severity("severity")}

	for i := 0; i < 10; i++ {
		err = sendEvent(wf, "test event", time.Now().Unix(), 0, source, tags, options...)
		if err != nil {
			log.Fatal(err)
			os.Exit(-1)
		}
	}

	// err = sendEvent(
	// 	senderP,
	// 	"Started container",
	// 	time.Now().Unix(), 0,
	// 	"gke-glaullon-default-pool-06d3ec34-c4qp",
	// 	map[string]string{
	// 		"namespace_name": " wavefront-collector",
	// 		"kind":           " Pod",
	// 		"reason":         " Started",
	// 		"component":      " kubelet",
	// 		"cluster":        " glaullon-deployment",
	// 	},
	// 	event.Type("Normal"),
	// )
	// if err != nil {
	// 	log.Fatal(err)
	// 	os.Exit(-1)
	// }

	// @Event 1571819354000 1571819354001 "Started container" type="Normal" host="gke-glaullon-default-pool-06d3ec34-c4qp" tag="namespace_name: wavefront-collector" tag="kind: Pod" tag="reason: Started" tag="component: kubelet" tag="cluster: glaullon-deployment"\n

	// for i := 0; i < 10; i++ {
	// 	sender.SendEvent(
	// 		"test event 1",
	// 		time.Now().Add(time.Second*time.Duration(-30)).Unix(),
	// 		time.Now().Unix(),
	// 		source,
	// 		[]string{"tag ", "tag 2"},
	// 		event.Details("Details"),
	// 		event.Type("Type"),
	// 		event.Severity("Severity"),
	// 	)
	// 	for c := 2; c < 25; c++ {
	// 		sender.SendEvent(fmt.Sprintf("test event %d", c), time.Now().Add(time.Second*time.Duration(-30)).Unix(), time.Now().Unix(), source, nil)
	// 	}
	// 	time.Sleep(time.Minute)
	// }
	time.Sleep(5 * time.Minute)
	wf.Flush()
	wf.Close()
}

func sendEvent(sender wavefront.Client, name string, startMillis, endMillis int64, source string, tags map[string]string, setters ...event.Option) error {
	return sender.SendEvent(name, startMillis, endMillis, source, tags, setters...)
}
