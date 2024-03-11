package util

import (
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func TestRandomString(t *testing.T) {
	randomString := RandomString(10)
	assert.Equal(t, len(randomString), 10)
}

func TestRandomID(t *testing.T) {
	randomID := RandomID("pre")
	assert.Assert(t, strings.HasPrefix(randomID, "pre"))
}
