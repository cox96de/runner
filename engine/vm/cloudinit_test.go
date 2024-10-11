package vm

import (
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func Test_cloudInitUserData_Marshal(t *testing.T) {
	marshal, err := (&cloudInitUserData{
		RunCMD: [][]string{{"echo", "name"}},
	}).Marshal()
	assert.NilError(t, err)
	assert.Assert(t, strings.HasPrefix(marshal, "#cloud-config"))
}
