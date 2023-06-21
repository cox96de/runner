package util

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestMakeEnvPairs(t *testing.T) {
	envPairs := MakeEnvPairs(map[string]string{"a": "b"}, map[string]string{"c": "d"})
	assert.DeepEqual(t, envPairs, []string{"a=b", "c=d"})
}
