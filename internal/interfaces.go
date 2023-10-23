// Package internal offers helper interfaces that are internal to the Wavefront Go SDK.
// Interfaces within this package are not guaranteed to be backwards compatible between releases.
package internal

type Flusher interface {
	Flush() error
	FlushAll() error
	GetFailureCount() int64
	Start()
}

