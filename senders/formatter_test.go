package senders

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wavefronthq/wavefront-sdk-go/histogram"
)

var line string

func TestSanitizeInternal(t *testing.T) {
	assert.Equal(t, "\"hello\"", strconv.Quote(sanitizeInternal("hello")))
	assert.Equal(t, "\"hello-world\"", strconv.Quote(sanitizeInternal("hello world")))
	assert.Equal(t, "\"hello.world\"", strconv.Quote(sanitizeInternal("hello.world")))
	assert.Equal(t, "\"hello-world-\"", strconv.Quote(sanitizeInternal("hello\"world\"")))
	assert.Equal(t, "\"hello-world\"", strconv.Quote(sanitizeInternal("hello'world")))
	assert.Equal(t, "\"~component.heartbeat\"", strconv.Quote(sanitizeInternal("~component."+
		"heartbeat")))
	assert.Equal(t, "\"-component.heartbeat\"", strconv.Quote(sanitizeInternal("!component."+
		"heartbeat")))
	assert.Equal(t, "\"Δcomponent.heartbeat\"", strconv.Quote(sanitizeInternal("Δcomponent."+
		"heartbeat")))
	assert.Equal(t, "\"∆component.heartbeat\"", strconv.Quote(sanitizeInternal("∆component."+
		"heartbeat")))
	assert.Equal(t, "\"∆~component.heartbeat\"", strconv.Quote(sanitizeInternal("∆~component."+
		"heartbeat")))
}

func TestSanitizeValue(t *testing.T) {
	assert.Equal(t, "\"hello\"", sanitizeValue("hello"))
	assert.Equal(t, "\"hello world\"", sanitizeValue("hello world"))
	assert.Equal(t, "\"hello.world\"", sanitizeValue("hello.world"))
	assert.Equal(t, "\"hello\\\"world\\\"\"", sanitizeValue("hello\"world\""))
	assert.Equal(t, "\"hello'world\"", sanitizeValue("hello'world"))
	assert.Equal(t, "\"hello\\nworld\"", sanitizeValue("hello\nworld"))
}

func BenchmarkMetricLine(b *testing.B) {
	name := "foo.metric"
	value := 1.2
	ts := int64(1533529977)
	src := "test_source"
	tags := map[string]string{"env": "test"}

	var r string
	for n := 0; n < b.N; n++ {
		r, _ = metricLine(name, value, ts, src, tags, "")
	}
	line = r
}

func TestMetricLine(t *testing.T) {
	line, err := metricLine("foo.metric", 1.2, 1533529977, "test_source",
		map[string]string{"env": "test"}, "")
	expected := "\"foo.metric\" 1.2 1533529977 source=\"test_source\" \"env\"=\"test\"\n"
	assert.Nil(t, err)
	assert.Equal(t, expected, line)

	line, err = metricLine("foo.metric", 1.2, 1533529977, "",
		map[string]string{"env": "test"}, "default")
	expected = "\"foo.metric\" 1.2 1533529977 source=\"default\" \"env\"=\"test\"\n"
	assert.Nil(t, err)
	assert.Equal(t, expected, line)

	line, err = metricLine("foo.metric", 1.2, 1533529977, "1.2.3.4:8080",
		map[string]string{"env": "test"}, "default")
	expected = "\"foo.metric\" 1.2 1533529977 source=\"1.2.3.4:8080\" \"env\"=\"test\"\n"
	assert.Nil(t, err)
	assert.Equal(t, expected, line)
}

func BenchmarkHistogramLine(b *testing.B) {
	name := "request.latency"
	centroids := makeCentroids()
	hgs := map[histogram.Granularity]bool{histogram.MINUTE: true}
	ts := int64(1533529977)
	src := "test_source"
	tags := map[string]string{"env": "test"}

	var r string
	for n := 0; n < b.N; n++ {
		r, _ = histogramLine(name, centroids, hgs, ts, src, tags, "")
	}
	line = r
}

func TestHistogramLineCentroidsFormat(t *testing.T) {
	centroids := histogram.Centroids{
		{Value: 30.0, Count: 20},
		{Value: 5.1, Count: 10},
		{Value: 30.0, Count: 20},
		{Value: 5.1, Count: 10},
		{Value: 30.0, Count: 20},
	}

	line, err := histogramLine("request.latency", centroids, map[histogram.Granularity]bool{histogram.MINUTE: true},
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

	line, err := histogramLine("request.latency", centroids, map[histogram.Granularity]bool{histogram.MINUTE: true},
		1533529977, "test_source", map[string]string{"env": "test"}, "")
	expected := "!M 1533529977 #20 30 \"request.latency\" source=\"test_source\" \"env\"=\"test\"\n"
	assert.Nil(t, err)
	assert.Equal(t, expected, line)

	line, err = histogramLine("request.latency", centroids, map[histogram.Granularity]bool{histogram.MINUTE: true, histogram.HOUR: false},
		1533529977, "", map[string]string{"env": "test"}, "default")
	expected = "!M 1533529977 #20 30 \"request.latency\" source=\"default\" \"env\"=\"test\"\n"
	assert.Nil(t, err)
	assert.Equal(t, expected, line)

	line, err = histogramLine("request.latency", centroids, map[histogram.Granularity]bool{histogram.HOUR: true, histogram.MINUTE: false},
		1533529977, "", map[string]string{"env": "test"}, "default")
	expected = "!H 1533529977 #20 30 \"request.latency\" source=\"default\" \"env\"=\"test\"\n"
	assert.Nil(t, err)
	assert.Equal(t, expected, line)

	line, err = histogramLine("request.latency", centroids, map[histogram.Granularity]bool{histogram.DAY: true},
		1533529977, "", map[string]string{"env": "test"}, "default")
	expected = "!D 1533529977 #20 30 \"request.latency\" source=\"default\" \"env\"=\"test\"\n"
	assert.Nil(t, err)
	assert.Equal(t, expected, line)

	line, err = histogramLine("request.latency", centroids, map[histogram.Granularity]bool{histogram.MINUTE: true, histogram.HOUR: true, histogram.DAY: false},
		1533529977, "test_source", map[string]string{"env": "test"}, "")
	expected = "!M 1533529977 #20 30 \"request.latency\" source=\"test_source\" \"env\"=\"test\"\n" +
		"!H 1533529977 #20 30 \"request.latency\" source=\"test_source\" \"env\"=\"test\"\n"
	if len(line) != len(expected) {
		t.Errorf("lines don't match. expected: %s, actual: %s", expected, line)
	}
}

func BenchmarkSpanLine(b *testing.B) {
	name := "order.shirts"
	start := int64(1533531013)
	dur := int64(343500)
	src := "test_source"
	traceId := "7b3bf470-9456-11e8-9eb6-529269fb1459"

	var r string
	for n := 0; n < b.N; n++ {
		r, _ = spanLine(name, start, dur, src, traceId, traceId, []string{traceId}, nil, nil, nil, "")
	}
	line = r
}

func TestSpanLine(t *testing.T) {
	line, err := spanLine("order.shirts", 1533531013, 343500, "test_source",
		"7b3bf470-9456-11e8-9eb6-529269fb1459", "7b3bf470-9456-11e8-9eb6-529269fb1459",
		[]string{"7b3bf470-9456-11e8-9eb6-529269fb1458"}, nil, nil, nil, "")
	expected := "\"order.shirts\" source=\"test_source\" traceId=7b3bf470-9456-11e8-9eb6-529269fb1459" +
		" spanId=7b3bf470-9456-11e8-9eb6-529269fb1459 parent=7b3bf470-9456-11e8-9eb6-529269fb1458 1533531013 343500\n"
	assert.Nil(t, err)
	assert.Equal(t, expected, line)

	line, err = spanLine("order.shirts", 1533531013, 343500, "test_source",
		"7b3bf470-9456-11e8-9eb6-529269fb1459", "7b3bf470-9456-11e8-9eb6-529269fb1459", nil,
		[]string{"7b3bf470-9456-11e8-9eb6-529269fb1458"}, []SpanTag{{Key: "env", Value: "test"}}, nil, "")
	expected = "\"order.shirts\" source=\"test_source\" traceId=7b3bf470-9456-11e8-9eb6-529269fb1459" +
		" spanId=7b3bf470-9456-11e8-9eb6-529269fb1459 followsFrom=7b3bf470-9456-11e8-9eb6-529269fb1458 \"env\"=\"test\" 1533531013 343500\n"
	assert.Nil(t, err)
	assert.Equal(t, expected, line)

	line, err = spanLine("order.shirts", 1533531013, 343500, "test_source",
		"7b3bf470-9456-11e8-9eb6-529269fb1459", "7b3bf470-9456-11e8-9eb6-529269fb1459", nil,
		[]string{"7b3bf470-9456-11e8-9eb6-529269fb1458"},
		[]SpanTag{{Key: "env", Value: "test"}, {Key: "env", Value: "dev"}}, nil, "")
	expected = "\"order.shirts\" source=\"test_source\" traceId=7b3bf470-9456-11e8-9eb6-529269fb1459" +
		" spanId=7b3bf470-9456-11e8-9eb6-529269fb1459 followsFrom=7b3bf470-9456-11e8-9eb6-529269fb1458 \"env\"=\"test\" \"env\"=\"dev\" 1533531013 343500\n"
	assert.Nil(t, err)
	assert.Equal(t, expected, line)
}

func TestSpanLineErrors(t *testing.T) {
	uuid := "00000000-0000-0000-0000-000000000000"

	_, err := spanLine("", 0, 0, "", uuid, uuid, nil, nil, nil, nil, "")
	require.Error(t, err)
	assert.Equal(t, "span name cannot be empty", err.Error())

	_, err = spanLine("a_name", 0, 0, "00-00", "x", uuid, nil, nil, nil, nil, "")
	require.Error(t, err)
	assert.Equal(t, "traceId is not in UUID format: span=a_name traceId=x", err.Error())

	_, err = spanLine("a_name", 0, 0, "00-00", uuid, "x", nil, nil, nil, nil, "")
	require.Error(t, err)
	assert.Equal(t, "spanId is not in UUID format: span=a_name spanId=x", err.Error())

	_, err = spanLine("a_name", 0, 0, "a_source", uuid, uuid, nil, nil,
		[]SpanTag{{Key: "", Value: ""}}, nil, "")
	require.Error(t, err)
	assert.Equal(t, "tag keys cannot be empty: span=a_name", err.Error())

	_, err = spanLine("a_name", 0, 0, "a_source", uuid, uuid, nil, nil,
		[]SpanTag{{Key: "a_tag", Value: ""}}, nil, "")
	require.Error(t, err)
	assert.Equal(t, "tag values cannot be empty: span=a_name tag=a_tag", err.Error())
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

func TestVerifyUUID(tt *testing.T) {
	tt.Run("Good UUID 1", func(t *testing.T) {
		if isUUIDFormat("00112233-4455-6677-8899-aabbccddeeff") == false {
			t.Fail()
		}
	})
	tt.Run("Good UUID 2", func(t *testing.T) {
		if isUUIDFormat("AABBCCDD-EEFF-0011-2233-445566778899") == false {
			t.Fail()
		}
	})
	tt.Run("Bad UUID 1", func(t *testing.T) {
		if isUUIDFormat("00112233-4455-6677-8899-aabbccddee") == true {
			t.Fail()
		}
	})
	tt.Run("Bad UUID 2", func(t *testing.T) {
		if isUUIDFormat("00112233-445506677-8899-aabbccddeeff") == true {
			t.Fail()
		}
	})
	tt.Run("Bad UUID 3", func(t *testing.T) {
		if isUUIDFormat("00112233-44SS-6677-8899-aabbccddeeff") == true {
			t.Fail()
		}
	})
}
