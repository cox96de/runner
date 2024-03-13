package dispatch

import (
	"testing"

	"github.com/cox96de/runner/api"
	"gotest.tools/v3/assert"
)

func TestCheckJobStatus(t *testing.T) {
	check := CheckJobStatus(api.StatusPreparing, api.StatusRunning)
	assert.Assert(t, check == true)
	check = CheckJobStatus(api.StatusPreparing, api.StatusCreated)
	assert.Assert(t, check == false)
}

func TestCheckStepStatus(t *testing.T) {
	check := CheckStepStatus(api.StatusRunning, api.StatusRunning)
	assert.Assert(t, check == true)
	check = CheckStepStatus(api.StatusRunning, api.StatusCreated)
	assert.Assert(t, check == false)
}
