// Package senders provides functionality for sending data to Wavefront through
// the Wavefront proxy or via direct ingestion.
package senders

import (
	"github.com/wavefronthq/wavefront-sdk-go/wavefront"
)

// Sender Interface for sending metrics, distributions and spans to Wavefront
type Sender interface {
	wavefront.Client
}
