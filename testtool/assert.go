package testtool

import (
	"encoding/json"
	"os"
	"reflect"

	"gotest.tools/v3/assert"
)

//go:generate mockgen -destination mock/mockgen_tb.go -package mock . TestingT

type TestingT interface {
	assert.TestingT
	Helper()
}

// DeepEqualObject compares the given object with the object in the given file.
// It uses json as encoder and decoder.
func DeepEqualObject(t TestingT, got interface{}, expectPath string) {
	t.Helper()
	file, err := os.ReadFile(expectPath)
	if err != nil {
		if os.IsNotExist(err) {
			marshal, err := json.MarshalIndent(got, "", "  ")
			assert.NilError(t, err)
			err = os.WriteFile(expectPath, marshal, 0o644)
			assert.NilError(t, err)
		}
		assert.NilError(t, err)
	}
	typeOf := reflect.TypeOf(got)
	v := reflect.New(typeOf)
	value := v.Interface()
	err = json.Unmarshal(file, value)
	assert.NilError(t, err)
	assert.DeepEqual(t, got, reflect.Indirect(v).Interface())
}
