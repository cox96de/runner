package util

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestRandomInt(t *testing.T) {
	max := int64(100)
	for i := 0; i < 1000; i++ {
		n := RandomInt(0, max)
		if n < 0 || n >= max {
			t.Errorf("RandomInt(%d) = %d, out of range", max, n)
		}
	}
}

func TestRandomBytes(t *testing.T) {
	bytes := RandomBytes(10)
	assert.Equal(t, len(bytes), 10)
}
