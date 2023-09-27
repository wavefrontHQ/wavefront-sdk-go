package application_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-sdk-go/application"
)

func TestAppTagsEnv(t *testing.T) {
	assert.NoError(t, os.Setenv("app_label_1", "value_1"))
	assert.NoError(t, os.Setenv("app_label_2", "value_2"))
	assert.NoError(t, os.Setenv("label_3", "value_3"))

	appTags := application.New("app", "srv")

	assert.NoError(t, appTags.AddCustomTagsFromEnv("app_.*"))
	assert.NoError(t, appTags.AddCustomTagFromEnv("label_3", "app_3"))

	tags := appTags.Map()
	assert.Equal(t, "value_1", tags["app_label_1"])
	assert.Empty(t, tags["label_3"])
	assert.Equal(t, "value_3", tags["app_3"])

	assert.NotNil(t, appTags.AddCustomTagsFromEnv("ap\\p_.*"))
	assert.NotNil(t, appTags.AddCustomTagFromEnv("label_x", "app_3"))
}
