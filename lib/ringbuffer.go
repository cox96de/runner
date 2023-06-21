package lib

import (
	"io"
	"sync/atomic"

	"github.com/pkg/errors"
	"github.com/smallnest/ringbuffer"
)

// RingBuffer is a ring buffer with a Close() method.
type RingBuffer struct {
	r      *ringbuffer.RingBuffer
	closed atomic.Bool
}

// NewRingBuffer creates a new ring buffer with given size.
func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		closed: atomic.Bool{},
		r:      ringbuffer.New(size),
	}
}

// Close closes the ring buffer.
func (r *RingBuffer) Close() error {
	r.closed.Store(true)
	return nil
}

// Write writes data to the ring buffer.
func (r *RingBuffer) Write(p []byte) (n int, err error) {
	// TODO: block if buffer is full
	if r.closed.Load() {
		return 0, errors.Errorf("ring buffer is closed")
	}
	write, err := r.r.Write(p)
	if err == ringbuffer.ErrTooManyDataToWrite {
		return write, nil
	}
	return write, err
}

// Read reads data from the ring buffer.
func (r *RingBuffer) Read(p []byte) (n int, err error) {
	// TODO: block if buffer is empty
	n, err = r.r.Read(p)
	if n == 0 && err == ringbuffer.ErrIsEmpty {
		if r.closed.Load() {
			return n, io.EOF
		}
		return n, nil
	}
	return n, err
}
