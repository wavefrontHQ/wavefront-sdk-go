package senders

import (
	"strconv"
	"testing"

	"github.com/wavefronthq/wavefront-sdk-go/histogram"
)

var line string

func TestSanitizeInternal(t *testing.T) {
	assertEqual(t, "\"hello\"", strconv.Quote(sanitizeInternal("hello")))
	assertEqual(t, "\"hello-world\"", strconv.Quote(sanitizeInternal("hello world")))
	assertEqual(t, "\"hello.world\"", strconv.Quote(sanitizeInternal("hello.world")))
	assertEqual(t, "\"hello-world-\"", strconv.Quote(sanitizeInternal("hello\"world\"")))
	assertEqual(t, "\"hello-world\"", strconv.Quote(sanitizeInternal("hello'world")))
	assertEqual(t, "\"~component.heartbeat\"", strconv.Quote(sanitizeInternal("~component."+
		"heartbeat")))
	assertEqual(t, "\"-component.heartbeat\"", strconv.Quote(sanitizeInternal("!component."+
		"heartbeat")))
	assertEqual(t, "\"Δcomponent.heartbeat\"", strconv.Quote(sanitizeInternal("Δcomponent."+
		"heartbeat")))
	assertEqual(t, "\"∆component.heartbeat\"", strconv.Quote(sanitizeInternal("∆component."+
		"heartbeat")))
}

func TestSanitizeValue(t *testing.T) {
	assertEqual(t, "\"hello\"", sanitizeValue("hello"))
	assertEqual(t, "\"hello world\"", sanitizeValue("hello world"))
	assertEqual(t, "\"hello.world\"", sanitizeValue("hello.world"))
	assertEqual(t, "\"hello\\\"world\\\"\"", sanitizeValue("hello\"world\""))
	assertEqual(t, "\"hello'world\"", sanitizeValue("hello'world"))
	assertEqual(t, "\"hello\\nworld\"", sanitizeValue("hello\nworld"))
}

func BenchmarkMetricLine(b *testing.B) {
	name := "foo.metric"
	value := 1.2
	ts := int64(1533529977)
	src := "test_source"
	tags := map[string]string{"env": "test"}

	var r string
	for n := 0; n < b.N; n++ {
		r, _ = MetricLine(name, value, ts, src, tags, "")
	}
	line = r
}

func TestMetricLine(t *testing.T) {
	line, err := MetricLine("foo.metric", 1.2, 1533529977, "test_source",
		map[string]string{"env": "test"}, "")
	expected := "\"foo.metric\" 1.2 1533529977 source=\"test_source\" \"env\"=\"test\"\n"
	assertEquals(expected, line, err, t)

	line, err = MetricLine("foo.metric", 1.2, 1533529977, "",
		map[string]string{"env": "test"}, "default")
	expected = "\"foo.metric\" 1.2 1533529977 source=\"default\" \"env\"=\"test\"\n"
	assertEquals(expected, line, err, t)
}

func BenchmarkHistoLine(b *testing.B) {
	name := "request.latency"
	centroids := makeCentroids()
	hgs := map[histogram.Granularity]bool{histogram.MINUTE: true}
	ts := int64(1533529977)
	src := "test_source"
	tags := map[string]string{"env": "test"}

	var r string
	for n := 0; n < b.N; n++ {
		r, _ = HistoLine(name, centroids, hgs, ts, src, tags, "")
	}
	line = r
}

func TestHistoLine(t *testing.T) {
	centroids := makeCentroids()

	line, err := HistoLine("request.latency", centroids, map[histogram.Granularity]bool{histogram.MINUTE: true},
		1533529977, "test_source", map[string]string{"env": "test"}, "")
	expected := "!M 1533529977 #20 30 #10 5.1 \"request.latency\" source=\"test_source\" \"env\"=\"test\"\n"
	assertEquals(expected, line, err, t)

	line, err = HistoLine("request.latency", centroids, map[histogram.Granularity]bool{histogram.MINUTE: true, histogram.HOUR: false},
		1533529977, "", map[string]string{"env": "test"}, "default")
	expected = "!M 1533529977 #20 30 #10 5.1 \"request.latency\" source=\"default\" \"env\"=\"test\"\n"
	assertEquals(expected, line, err, t)

	line, err = HistoLine("request.latency", centroids, map[histogram.Granularity]bool{histogram.HOUR: true, histogram.MINUTE: false},
		1533529977, "", map[string]string{"env": "test"}, "default")
	expected = "!H 1533529977 #20 30 #10 5.1 \"request.latency\" source=\"default\" \"env\"=\"test\"\n"
	assertEquals(expected, line, err, t)

	line, err = HistoLine("request.latency", centroids, map[histogram.Granularity]bool{histogram.DAY: true},
		1533529977, "", map[string]string{"env": "test"}, "default")
	expected = "!D 1533529977 #20 30 #10 5.1 \"request.latency\" source=\"default\" \"env\"=\"test\"\n"
	assertEquals(expected, line, err, t)

	line, err = HistoLine("request.latency", centroids, map[histogram.Granularity]bool{histogram.MINUTE: true, histogram.HOUR: true, histogram.DAY: false},
		1533529977, "test_source", map[string]string{"env": "test"}, "")
	expected = "!M 1533529977 #20 30 #10 5.1 \"request.latency\" source=\"test_source\" \"env\"=\"test\"\n" +
		"!H 1533529977 #20 30 #10 5.1 \"request.latency\" source=\"test_source\" \"env\"=\"test\"\n"
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
		r, _ = SpanLine(name, start, dur, src, traceId, traceId, []string{traceId}, nil, nil, nil, "")
	}
	line = r
}

func TestSpanLine(t *testing.T) {
	line, err := SpanLine("order.shirts", 1533531013, 343500, "test_source",
		"7b3bf470-9456-11e8-9eb6-529269fb1459", "7b3bf470-9456-11e8-9eb6-529269fb1459",
		[]string{"7b3bf470-9456-11e8-9eb6-529269fb1458"}, nil, nil, nil, "")
	expected := "\"order.shirts\" source=\"test_source\" traceId=7b3bf470-9456-11e8-9eb6-529269fb1459" +
		" spanId=7b3bf470-9456-11e8-9eb6-529269fb1459 parent=7b3bf470-9456-11e8-9eb6-529269fb1458 1533531013 343500\n"
	assertEquals(expected, line, err, t)

	line, err = SpanLine("order.shirts", 1533531013, 343500, "test_source",
		"7b3bf470-9456-11e8-9eb6-529269fb1459", "7b3bf470-9456-11e8-9eb6-529269fb1459", nil,
		[]string{"7b3bf470-9456-11e8-9eb6-529269fb1458"}, []SpanTag{{Key: "env", Value: "test"}}, nil, "")
	expected = "\"order.shirts\" source=\"test_source\" traceId=7b3bf470-9456-11e8-9eb6-529269fb1459" +
		" spanId=7b3bf470-9456-11e8-9eb6-529269fb1459 followsFrom=7b3bf470-9456-11e8-9eb6-529269fb1458 \"env\"=\"test\" 1533531013 343500\n"
	assertEquals(expected, line, err, t)

	line, err = SpanLine("order.shirts", 1533531013, 343500, "test_source",
		"7b3bf470-9456-11e8-9eb6-529269fb1459", "7b3bf470-9456-11e8-9eb6-529269fb1459", nil,
		[]string{"7b3bf470-9456-11e8-9eb6-529269fb1458"},
		[]SpanTag{{Key: "env", Value: "test"}, {Key: "env", Value: "dev"}}, nil, "")
	expected = "\"order.shirts\" source=\"test_source\" traceId=7b3bf470-9456-11e8-9eb6-529269fb1459" +
		" spanId=7b3bf470-9456-11e8-9eb6-529269fb1459 followsFrom=7b3bf470-9456-11e8-9eb6-529269fb1458 \"env\"=\"test\" \"env\"=\"dev\" 1533531013 343500\n"
	assertEquals(expected, line, err, t)
}

func assertEquals(expected, actual string, err error, t *testing.T) {
	if err != nil {
		t.Error(err)
	}
	if actual != expected {
		t.Errorf("lines don't match.\n expected: %s\n actual: %s", expected, actual)
	}
}

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%s - %v != %v", "", a, b)
	}
}

func makeCentroids() []histogram.Centroid {
	centroids := []histogram.Centroid{
		{
			Value: 30.0,
			Count: 20,
		},
		{
			Value: 5.1,
			Count: 10,
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
