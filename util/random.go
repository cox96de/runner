package util

import (
	"crypto/rand"
	mrand "math/rand"
)

// RandomBytes returns a slice of random bytes of the given size.
func RandomBytes(size int) []byte {
	bytes := make([]byte, size)
	_, _ = rand.Reader.Read(bytes)
	return bytes
}

// RandomInt returns a random int64 in [0, max).
func RandomInt(max int64) int64 {
	return mrand.Int63n(max)
}
