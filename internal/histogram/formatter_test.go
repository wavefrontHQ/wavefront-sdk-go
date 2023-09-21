package histogram

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-sdk-go/histogram"
)

var benchmarkLine string

func BenchmarkHistogramLine(b *testing.B) {
	name := "request.latency"
	centroids := makeCentroids()
	hgs := map[histogram.Granularity]bool{histogram.MINUTE: true}
	ts := int64(1533529977)
	src := "test_source"
	tags := map[string]string{"env": "test"}

	var r string
	for n := 0; n < b.N; n++ {
		r, _ = Line(name, centroids, hgs, ts, src, tags, "")
	}
	benchmarkLine = r
}

func TestHistogramLineCentroidsFormat(t *testing.T) {
	centroids := histogram.Centroids{
		{Value: 30.0, Count: 20},
		{Value: 5.1, Count: 10},
		{Value: 30.0, Count: 20},
		{Value: 5.1, Count: 10},
		{Value: 30.0, Count: 20},
	}

	line, err := Line("request.latency", centroids, map[histogram.Granularity]bool{histogram.MINUTE: true},
		1533529977, "test_source", map[string]string{"env": "test"}, "")

	assert.Nil(t, err)
	expected := []string{
		"!M 1533529977 #60 30 #20 5.1 \"request.latency\" source=\"test_source\" \"env\"=\"test\"\n",
		"!M 1533529977 #20 5.1 #60 30 \"request.latency\" source=\"test_source\" \"env\"=\"test\"\n",
	}
	ok := false
	for _, exp := range expected {
		if assert.ObjectsAreEqual(exp, line) {
			ok = true
		}
	}
	if !ok {
		assert.Equal(t, expected[0], line)
		assert.Equal(t, expected[1], line)
	}
}

func TestHistogramLine(t *testing.T) {
	centroids := makeCentroids()

	line, err := Line("request.latency", centroids, map[histogram.Granularity]bool{histogram.MINUTE: true},
		1533529977, "test_source", map[string]string{"env": "test"}, "")
	expected := "!M 1533529977 #20 30 \"request.latency\" source=\"test_source\" \"env\"=\"test\"\n"
	assert.NoError(t, err)
	assert.Equal(t, expected, line)

	line, err = Line("request.latency", centroids, map[histogram.Granularity]bool{histogram.MINUTE: true, histogram.HOUR: false},
		1533529977, "", map[string]string{"env": "test"}, "default")
	expected = "!M 1533529977 #20 30 \"request.latency\" source=\"default\" \"env\"=\"test\"\n"
	assert.NoError(t, err)
	assert.Equal(t, expected, line)

	line, err = Line("request.latency", centroids, map[histogram.Granularity]bool{histogram.HOUR: true, histogram.MINUTE: false},
		1533529977, "", map[string]string{"env": "test"}, "default")
	expected = "!H 1533529977 #20 30 \"request.latency\" source=\"default\" \"env\"=\"test\"\n"
	assert.NoError(t, err)
	assert.Equal(t, expected, line)

	line, err = Line("request.latency", centroids, map[histogram.Granularity]bool{histogram.DAY: true},
		1533529977, "", map[string]string{"env": "test"}, "default")
	expected = "!D 1533529977 #20 30 \"request.latency\" source=\"default\" \"env\"=\"test\"\n"
	assert.NoError(t, err)
	assert.Equal(t, expected, line)

	line, err = Line("request.latency", centroids, map[histogram.Granularity]bool{histogram.MINUTE: true, histogram.HOUR: true, histogram.DAY: false},
		1533529977, "test_source", map[string]string{"env": "test"}, "")

	assert.NoError(t, err)
	assert.ElementsMatch(t, strings.Split(line, "\n")[0:2], []string{
		"!M 1533529977 #20 30 \"request.latency\" source=\"test_source\" \"env\"=\"test\"",
		"!H 1533529977 #20 30 \"request.latency\" source=\"test_source\" \"env\"=\"test\"",
	})

}

func makeCentroids() []histogram.Centroid {
	centroids := []histogram.Centroid{
		{
			Value: 30.0,
			Count: 20,
		},
	}
	return centroids
}
