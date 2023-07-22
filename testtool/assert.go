package testtool

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"gotest.tools/v3/assert"
)

//go:generate mockgen -destination mock/mockgen_tb.go -package mock . TestingT

type TestingT interface {
	Logf(format string, args ...any)
	FailNow()
	Helper()
}

// DeepEqualObject compares the given object with the object in the given file.
// It uses json as encoder and decoder.
func DeepEqualObject(t *testing.T, got interface{}, expectPath string) {
	t.Helper()
	AssertObject(t, got, expectPath, func(got, expect interface{}) {
		assert.DeepEqual(t, got, expect)
	})
}

// AssertObject compares the given object with the object in the given file.
// It uses json as encoder and decoder.
func AssertObject(t TestingT, got interface{}, expectPath string,
	assert func(got, expect interface{}),
) {
	t.Helper()
	file, err := os.ReadFile(expectPath)
	if err != nil {
		if os.IsNotExist(err) {
			marshal, err := json.MarshalIndent(got, "", "  ")
			nilError(t, err)
			err = os.WriteFile(expectPath, marshal, 0o644)
			nilError(t, err)
		}
		nilError(t, err)
	}
	typeOf := reflect.TypeOf(got)
	v := reflect.New(typeOf)
	value := v.Interface()
	err = json.Unmarshal(file, value)
	nilError(t, err)
	assert(got, reflect.Indirect(v).Interface())
}

func nilError(t TestingT, err error) {
	if err != nil {
		t.Helper()
		t.Logf("error: %v", err)
		t.FailNow()
	}
}
