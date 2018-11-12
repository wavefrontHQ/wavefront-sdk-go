package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	metrics "github.com/wavefronthq/wavefront-sdk-go/metrics"
	wavefront "github.com/wavefronthq/wavefront-sdk-go/senders"
)

func main() {
	directCfg := &wavefront.DirectConfiguration{
		Server: "https://nimba.wavefront.com",          // your Wavefront instance URL
		Token:  "a87886d5-899d-4a23-98c0-2f3538ed007a", // API token with direct ingestion permission
	}

	sender, err := wavefront.NewDirectSender(directCfg)
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}

	log.Print("sender ready")
	h := metrics.NewHistogram("glaullon.histogram", "glaullon_histogram_test", map[string]string{"tag1": "tag"})
	h2 := metrics.NewHistogram("glaullon.histogram", "glaullon_histogram_test", map[string]string{"tag1": "tag"})
	h2.Granularity = wavefront.HOUR

	const delay = 5 * time.Second
	ticker := time.NewTicker(delay)

	go func() {
		for range ticker.C {
			h.Report(sender)
			h2.Report(sender)
		}
	}()

	for true {
		// for i := 0; i < 1000; i++ {
		h.Add(rand.Float64())
		h2.Add(rand.Float64())
		// }
		time.Sleep(time.Second)
	}
}
