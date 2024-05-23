package lib

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestBuildStarlarkValue(t *testing.T) {
	type Gostruct struct {
		Int               int
		Int32             int32
		Float32           float32
		MapStringString   map[string]string
		MapStringGostruct map[string]*Gostruct
		ListString        []string
		Ptr               *int
		Tag               string `json:"tag"`
	}
	a := &Gostruct{
		Int:     1,
		Int32:   2,
		Float32: 3.0,
		MapStringString: map[string]string{
			"key": "value",
		},
		MapStringGostruct: map[string]*Gostruct{"key": {
			Int: 10,
		}},
		ListString: []string{"a", "b"},
		Tag:        "tag",
	}

	value, err := ConvertToStarlarkValue(a)
	assert.NilError(t, err)
	assert.Assert(t, value != nil)
	// Make it easy to assert.
	assert.Equal(t, value.String(), `struct(Float32 = 3.0, Int = 1, Int32 = 2, ListString = ["a", "b"], MapStringGostruct = {"key": struct(Float32 = 0.0, Int = 10, Int32 = 0, ListString = [], MapStringGostruct = {}, MapStringString = {}, tag = "")}, MapStringString = {"key": "value"}, tag = "tag")`)
}
