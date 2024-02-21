package api

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestConvertTime(t *testing.T) {
	timestamp := ConvertTime(time.Now())
	assert.Assert(t, timestamp != nil)
	var pt *time.Time
	convertTime := ConvertTime(pt)
	assert.Assert(t, convertTime == nil)
}
