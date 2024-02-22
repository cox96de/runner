package api

import (
	"testing"
	"time"

	"github.com/samber/lo"

	"gotest.tools/v3/assert"
)

func TestConvertTime(t *testing.T) {
	now := time.Now()
	timestamp := ConvertTime(now)
	assert.Assert(t, timestamp.AsTime().Equal(now))
	var pt *time.Time
	convertTime := ConvertTime(pt)
	assert.Assert(t, convertTime == nil)

	convertTime = ConvertTime(lo.ToPtr(now))
	assert.Assert(t, convertTime.AsTime().Equal(now))
}
