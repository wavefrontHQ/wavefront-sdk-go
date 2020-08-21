package histogram

import (
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	h := New(MaxBins(3), GranularityOption(MINUTE), TimeSupplier(c.Now))

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
	centroids := Centroids{
		{Value: 30.0, Count: 20},
		{Value: 5.1, Count: 10},
		{Value: 30.0, Count: 20},
		{Value: 5.1, Count: 10},
		{Value: 30.0, Count: 20},
	}

	centroidsExp := Centroids{
		{Value: 5.1, Count: 20},
		{Value: 30.0, Count: 60},
	}

	vals := centroids.Compact()
	sort.Sort(vals)

	assert.Equal(t, centroidsExp, vals, "Error on Centroids.Compact()")
}

func (a Centroids) Len() int           { return len(a) }
func (a Centroids) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Centroids) Less(i, j int) bool { return a[i].Value < a[j].Value }
