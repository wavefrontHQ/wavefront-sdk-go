package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalculateNewTickerInterval(t *testing.T) {
	fallback := 1 * time.Second
	assert.Equal(t, 999999820*time.Second, calculateNewTickerInterval(1000000000*time.Second, fallback))
	assert.Equal(t, 569*time.Second, calculateNewTickerInterval(599*time.Second, fallback))
	assert.Equal(t, 1*time.Second, calculateNewTickerInterval(3*time.Second, fallback))
	assert.Equal(t, 1*time.Second, calculateNewTickerInterval(1*time.Second, fallback))
	assert.Equal(t, 1*time.Second, calculateNewTickerInterval(0*time.Second, fallback))
	assert.Equal(t, 1*time.Second, calculateNewTickerInterval(-180*time.Second, fallback))

	fallback = 60 * time.Second
	assert.Equal(t, 999999820*time.Second, calculateNewTickerInterval(1000000000*time.Second, fallback))
	assert.Equal(t, 569*time.Second, calculateNewTickerInterval(599*time.Second, fallback))
	assert.Equal(t, 60*time.Second, calculateNewTickerInterval(3*time.Second, fallback))
	assert.Equal(t, 60*time.Second, calculateNewTickerInterval(1*time.Second, fallback))
	assert.Equal(t, 60*time.Second, calculateNewTickerInterval(0*time.Second, fallback))
	assert.Equal(t, 60*time.Second, calculateNewTickerInterval(-180*time.Second, fallback))
}
