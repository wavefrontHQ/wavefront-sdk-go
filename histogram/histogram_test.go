package histogram_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
	"github.com/wavefronthq/wavefront-sdk-go/senders"
)

type clock struct {
	currTime time.Time
}

func (c *clock) Now() time.Time {
	return c.currTime
}

func (c *clock) Add(d time.Duration) {
	c.currTime = c.currTime.Add(d)
}

func TestHistogram(t *testing.T) {
	c := &clock{currTime: time.Now()}
	h := histogram.New(histogram.MaxBins(3), histogram.GranularityOption(histogram.MINUTE), histogram.TimeSupplier(c.Now))

	for i := 0; i < 5; i++ {
		for i := 0; i < 1000; i++ {
			h.Update(rand.Float64())
		}
		c.Add(61 * time.Second)
	}

	distributions := h.Distributions()
	assert.Equal(t, len(distributions), 3, "Error on distributions number")

	for _, distribution := range distributions {
		count := 0
		for _, centroid := range distribution.Centroids {
			count += centroid.Count
		}
		assert.Equal(t, count, 1000, "Error on centroids count")
	}

	distributions = h.Distributions()
	assert.Equal(t, len(distributions), 0, "Error on distributions number")
}

func TestCompactHistoLine(t *testing.T) {
	centroids := histogram.Centroids{
		{Value: 30.0, Count: 20},
		{Value: 30.0, Count: 20},
		{Value: 30.0, Count: 20},
		{Value: 5.1, Count: 10},
	}

	line, err := senders.HistoLine("request.latency", centroids, map[histogram.Granularity]bool{histogram.MINUTE: true},
		1533529977, "test_source", map[string]string{"env": "test"}, "")
	expected := "!M 1533529977 #10 5.1 #60 30 \"request.latency\" source=\"test_source\" \"env\"=\"test\"\n"
	assert.Nil(t, err)
	assert.Equal(t, expected, line, "Error on histogram.Centroids.Compact()")
}
