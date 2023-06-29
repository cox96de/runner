package util

import (
	"time"

	"golang.org/x/exp/rand"
)

const defaultCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func init() {
	rand.Seed(uint64(time.Now().UnixNano()))
}

// RandomString returns a random string with the given length.
func RandomString(length int) string {
	return RandomStringFromCharset(length, defaultCharset)
}

func RandomStringFromCharset(length int, charset string) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}
