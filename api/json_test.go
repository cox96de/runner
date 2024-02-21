package api

import (
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"
)

func Test_status_UnmarshalJSON(t *testing.T) {
	assert.Assert(t, len(Status_name) == len(status2str)+1)
	s := StatusQueued
	marshal, err := json.Marshal(s)
	assert.NilError(t, err)
	assert.Equal(t, string(marshal), "\"queued\"")
	var s2 Status
	err = json.Unmarshal(marshal, &s2)
	assert.NilError(t, err)
	assert.Equal(t, s, s2)
}
