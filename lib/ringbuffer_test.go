package lib

import (
	"io"
	"testing"

	"github.com/cox96de/runner/util"
	"gotest.tools/v3/assert"
)

func TestNewRingBuffer(t *testing.T) {
	t.Run("read_write", func(t *testing.T) {
		data := util.RandomBytes(5000)
		rb := NewRingBuffer(1024)
		go func() {
			data := data
			for {
				randInt := util.RandomInt(512)
				if randInt >= int64(len(data)) {
					_, err := rb.Write(data[:])
					assert.NilError(t, err)
					break
				}
				n, err := rb.Write(data[:randInt])
				assert.NilError(t, err, "write %d bytes", randInt)
				data = data[n:]
			}
			err := rb.Close()
			assert.NilError(t, err)
		}()
		all, err := io.ReadAll(rb)
		assert.NilError(t, err)
		assert.DeepEqual(t, all, data)
	})
	t.Run("close", func(t *testing.T) {
		rb := NewRingBuffer(512)
		err := rb.Close()
		assert.NilError(t, err)
		n, err := rb.Write(util.RandomBytes(10))
		assert.Equal(t, n, 0)
		assert.ErrorContains(t, err, "ring buffer is closed")
	})
	t.Run("exceed_buffer_size", func(t *testing.T) {
		bufSize := 512
		rb := NewRingBuffer(bufSize)
		testdata := util.RandomBytes(513)
		n, err := rb.Write(testdata)
		assert.Equal(t, n, bufSize)
		assert.NilError(t, err)
		_ = rb.Close()
		all, err := io.ReadAll(rb)
		assert.NilError(t, err)
		assert.DeepEqual(t, all, testdata[:bufSize])
	})
}
