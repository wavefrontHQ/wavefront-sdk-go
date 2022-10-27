package internal

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestSanitizeInternal(t *testing.T) {
	assert.Equal(t, "\"hello\"", strconv.Quote(Sanitize("hello")))
	assert.Equal(t, "\"hello-world\"", strconv.Quote(Sanitize("hello world")))
	assert.Equal(t, "\"hello.world\"", strconv.Quote(Sanitize("hello.world")))
	assert.Equal(t, "\"hello-world-\"", strconv.Quote(Sanitize("hello\"world\"")))
	assert.Equal(t, "\"hello-world\"", strconv.Quote(Sanitize("hello'world")))
	assert.Equal(t, "\"~component.heartbeat\"", strconv.Quote(Sanitize("~component."+
		"heartbeat")))
	assert.Equal(t, "\"-component.heartbeat\"", strconv.Quote(Sanitize("!component."+
		"heartbeat")))
	assert.Equal(t, "\"Δcomponent.heartbeat\"", strconv.Quote(Sanitize("Δcomponent."+
		"heartbeat")))
	assert.Equal(t, "\"∆component.heartbeat\"", strconv.Quote(Sanitize("∆component."+
		"heartbeat")))
	assert.Equal(t, "\"∆~component.heartbeat\"", strconv.Quote(Sanitize("∆~component."+
		"heartbeat")))
}

func TestSanitizeValue(t *testing.T) {
	assert.Equal(t, "\"hello\"", SanitizeValue("hello"))
	assert.Equal(t, "\"hello world\"", SanitizeValue("hello world"))
	assert.Equal(t, "\"hello.world\"", SanitizeValue("hello.world"))
	assert.Equal(t, "\"hello\\\"world\\\"\"", SanitizeValue("hello\"world\""))
	assert.Equal(t, "\"hello'world\"", SanitizeValue("hello'world"))
	assert.Equal(t, "\"hello\\nworld\"", SanitizeValue("hello\nworld"))
}
