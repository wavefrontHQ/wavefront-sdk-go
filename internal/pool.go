package internal

import (
	"sync"
)

var buffers *sync.Pool

func init() {
	buffers = &sync.Pool{
		New: func() interface{} {
			return new(StringBuilder)
		},
	}
}

// GetBuffer fetches a buffers from the pool
func GetBuffer() *StringBuilder {
	return buffers.Get().(*StringBuilder)
}

// PutBuffer returns a buffers to the pool
func PutBuffer(buf *StringBuilder) {
	buf.Reset()
	buffers.Put(buf)
}
