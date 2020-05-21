// Package application provides functionality for encapsulating details about your Go application.
// The application details can be reported to Wavefront in the form of tags.
package application

import (
	"log"
	"os"
	"regexp"
	"strings"
)

// Tags Encapsulates application details
type Tags struct {
	Application string
	Service     string
	Cluster     string
	Shard       string
	CustomTags  map[string]string
}

// New creates a new application Tags with application and service name
func New(application, service string) Tags {
	return Tags{
		Application: application,
		Service:     service,
		Cluster:     "none",
		Shard:       "none",
		CustomTags:  make(map[string]string, 0),
	}
}

// Map with all values
func (app *Tags) Map() map[string]string {
	allTags := make(map[string]string)
	allTags["application"] = app.Application
	allTags["service"] = app.Service
	allTags["cluster"] = app.Cluster
	allTags["shard"] = app.Shard

	for k, v := range app.CustomTags {
		allTags[k] = v
	}
	return allTags
}

// AddCustomTagsFromEnv set additional custom tags from environment variables that match the given regex.
func (app *Tags) AddCustomTagsFromEnv(regx string) {
	r, err := regexp.Compile(regx)
	if err != nil {
		log.Printf("Error creating custom tags: %v\n", err)
		return
	}

	env := os.Environ()
	for _, envVar := range env {
		k := strings.Split(envVar, "=")[0]
		if r.Match([]byte(k)) {
			v := os.Getenv(k)
			if len(v) > 0 {
				app.CustomTags[k] = v
			}
		}
	}
}

// AddCustomTagFromEnv Set a custom tag from the given environment variable.
func (app *Tags) AddCustomTagFromEnv(varName, tag string) {
	v := os.Getenv(varName)
	if len(v) > 0 {
		app.CustomTags[tag] = v
	}
}
