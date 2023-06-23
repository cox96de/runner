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
	// writeChan is used to notify the reader that there is data to read.
	writeChan chan struct{}
	// readChan is used to notify the writer that there is space to write.
	readChan chan struct{}
}

// NewRingBuffer creates a new ring buffer with given size.
func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		closed:    atomic.Bool{},
		r:         ringbuffer.New(size),
		writeChan: make(chan struct{}, 1),
		readChan:  make(chan struct{}, 1),
	}
}

// Close closes the ring buffer.
func (r *RingBuffer) Close() error {
	r.closed.Store(true)
	// Might a routine be blocked in Read().
	close(r.writeChan)
	return nil
}

// Write writes data to the ring buffer.
func (r *RingBuffer) Write(p []byte) (n int, err error) {
	if r.closed.Load() {
		return 0, errors.Errorf("ring buffer is closed")
	}
	if len(p) > r.r.Capacity() {
		p = p[:r.r.Capacity()]
	}
	for {
		write, err := r.r.Write(p)
		if write == 0 && (err == ringbuffer.ErrIsFull || err == ringbuffer.ErrTooManyDataToWrite) {
			<-r.readChan
			continue
		}
		notify(r.writeChan)
		return write, nil
	}
}

// Read reads data from the ring buffer.
func (r *RingBuffer) Read(p []byte) (n int, err error) {
	for {
		n, err = r.r.Read(p)
		if n == 0 && err == ringbuffer.ErrIsEmpty {
			if r.closed.Load() {
				return n, io.EOF
			}
			<-r.writeChan
			continue
		}
		notify(r.readChan)
		return n, err
	}
}

func notify(c chan<- struct{}) {
	select {
	case c <- struct{}{}:
	default:

	}
}
