package span

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var line string

func BenchmarkSpanLine(b *testing.B) {
	name := "order.shirts"
	start := int64(1533531013)
	dur := int64(343500)
	src := "test_source"
	traceID := "7b3bf470-9456-11e8-9eb6-529269fb1459"

	var r string
	for n := 0; n < b.N; n++ {
		r, _ = Line(name, start, dur, src, traceID, traceID, []string{traceID}, nil, nil, nil, "")
	}
	line = r
}

func TestSpanLine(t *testing.T) {
	line, err := Line("order.shirts", 1533531013, 343500, "test_source",
		"7b3bf470-9456-11e8-9eb6-529269fb1459", "7b3bf470-9456-11e8-9eb6-529269fb1459",
		[]string{"7b3bf470-9456-11e8-9eb6-529269fb1458"}, nil, nil, nil, "")
	expected := "\"order.shirts\" source=\"test_source\" traceId=7b3bf470-9456-11e8-9eb6-529269fb1459" +
		" spanId=7b3bf470-9456-11e8-9eb6-529269fb1459 parent=7b3bf470-9456-11e8-9eb6-529269fb1458 1533531013 343500\n"
	assert.Nil(t, err)
	assert.Equal(t, expected, line)

	line, err = Line("order.shirts", 1533531013, 343500, "test_source",
		"7b3bf470-9456-11e8-9eb6-529269fb1459", "7b3bf470-9456-11e8-9eb6-529269fb1459", nil,
		[]string{"7b3bf470-9456-11e8-9eb6-529269fb1458"}, []Tag{{Key: "env", Value: "test"}}, nil, "")
	expected = "\"order.shirts\" source=\"test_source\" traceId=7b3bf470-9456-11e8-9eb6-529269fb1459" +
		" spanId=7b3bf470-9456-11e8-9eb6-529269fb1459 followsFrom=7b3bf470-9456-11e8-9eb6-529269fb1458 \"env\"=\"test\" 1533531013 343500\n"
	assert.Nil(t, err)
	assert.Equal(t, expected, line)

	line, err = Line("order.shirts", 1533531013, 343500, "test_source",
		"7b3bf470-9456-11e8-9eb6-529269fb1459", "7b3bf470-9456-11e8-9eb6-529269fb1459", nil,
		[]string{"7b3bf470-9456-11e8-9eb6-529269fb1458"},
		[]Tag{{Key: "env", Value: "test"}, {Key: "env", Value: "dev"}}, nil, "")
	expected = "\"order.shirts\" source=\"test_source\" traceId=7b3bf470-9456-11e8-9eb6-529269fb1459" +
		" spanId=7b3bf470-9456-11e8-9eb6-529269fb1459 followsFrom=7b3bf470-9456-11e8-9eb6-529269fb1458 \"env\"=\"test\" \"env\"=\"dev\" 1533531013 343500\n"
	assert.Nil(t, err)
	assert.Equal(t, expected, line)
}

func TestSpanLineErrors(t *testing.T) {
	uuid := "00000000-0000-0000-0000-000000000000"

	_, err := Line("", 0, 0, "", uuid, uuid, nil, nil, nil, nil, "")
	require.Error(t, err)
	assert.Equal(t, "span name cannot be empty", err.Error())

	_, err = Line("a_name", 0, 0, "00-00", "x", uuid, nil, nil, nil, nil, "")
	require.Error(t, err)
	assert.Equal(t, "traceId is not in UUID format: span=a_name traceId=x", err.Error())

	_, err = Line("a_name", 0, 0, "00-00", uuid, "x", nil, nil, nil, nil, "")
	require.Error(t, err)
	assert.Equal(t, "spanId is not in UUID format: span=a_name spanId=x", err.Error())

	_, err = Line("a_name", 0, 0, "a_source", uuid, uuid, nil, nil,
		[]Tag{{Key: "", Value: ""}}, nil, "")
	require.Error(t, err)
	assert.Equal(t, "tag keys cannot be empty: span=a_name", err.Error())

	_, err = Line("a_name", 0, 0, "a_source", uuid, uuid, nil, nil,
		[]Tag{{Key: "a_tag", Value: ""}}, nil, "")
	require.Error(t, err)
	assert.Equal(t, "tag values cannot be empty: span=a_name tag=a_tag", err.Error())
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
