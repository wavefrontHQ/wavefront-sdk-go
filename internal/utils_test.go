package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSemVer(t *testing.T) {

	sdkVersion, e := GetSemVer("")
	assert.NoError(t, e)
	assert.Equal(t, sdkVersion, float64(0))

	sdkVersion, e = GetSemVer("1.1.0")
	assert.NoError(t, e)
	assert.Equal(t, sdkVersion, 1.0100)

	sdkVersion, e = GetSemVer("1.1.0-SNAPSHOT")
	assert.NoError(t, e)
	assert.Equal(t, sdkVersion, 1.0100)

	sdkVersion, e = GetSemVer("1.1.1")
	assert.NoError(t, e)
	assert.Equal(t, sdkVersion, 1.0101)

	sdkVersion, e = GetSemVer("1.10.1")
	assert.NoError(t, e)
	assert.Equal(t, sdkVersion, 1.1001)

	sdkVersion, e = GetSemVer("1.1.10")
	assert.NoError(t, e)
	assert.Equal(t, sdkVersion, 1.0110)

	sdkVersion, e = GetSemVer("1.0.1")
	assert.NoError(t, e)
	assert.Equal(t, sdkVersion, 1.0001)

	sdkVersion, e = GetSemVer("1.0.10")
	assert.NoError(t, e)
	assert.Equal(t, sdkVersion, 1.0010)

	sdkVersion, e = GetSemVer("1.10.10")
	assert.NoError(t, e)
	assert.Equal(t, sdkVersion, 1.1010)
}
