package dispatch

import (
	"testing"

	"github.com/cox96de/runner/api"
	"gotest.tools/v3/assert"
)

func TestCheckStatus(t *testing.T) {
	check := CheckStatus(api.StatusPreparing, api.StatusRunning)
	assert.Assert(t, check == true)
	check = CheckStatus(api.StatusPreparing, api.StatusCreated)
	assert.Assert(t, check == false)
}
