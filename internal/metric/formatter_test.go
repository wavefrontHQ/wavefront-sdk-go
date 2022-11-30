package metric

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var line string

func BenchmarkMetricLine(b *testing.B) {
	name := "foo.metric"
	value := 1.2
	ts := int64(1533529977)
	src := "test_source"
	tags := map[string]string{"env": "test"}

	var r string
	for n := 0; n < b.N; n++ {
		r, _ = Line(name, value, ts, src, tags, "")
	}
	line = r
}

func TestMetricLine(t *testing.T) {
	line, err := Line("foo.metric", 1.2, 1533529977, "test_source",
		map[string]string{"env": "test"}, "")
	expected := "\"foo.metric\" 1.2 1533529977 source=\"test_source\" \"env\"=\"test\"\n"
	assert.Nil(t, err)
	assert.Equal(t, expected, line)

	line, err = Line("foo.metric", 1.2, 1533529977, "",
		map[string]string{"env": "test"}, "default")
	expected = "\"foo.metric\" 1.2 1533529977 source=\"default\" \"env\"=\"test\"\n"
	assert.Nil(t, err)
	assert.Equal(t, expected, line)

	line, err = Line("foo.metric", 1.2, 1533529977, "1.2.3.4:8080",
		map[string]string{"env": "test"}, "default")
	expected = "\"foo.metric\" 1.2 1533529977 source=\"1.2.3.4:8080\" \"env\"=\"test\"\n"
	assert.Nil(t, err)
	assert.Equal(t, expected, line)
}
