package util

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestStringError_Error(t *testing.T) {
	const a = StringError("a")
	is := errors.Is(a, a)
	assert.Truef(t, is, "expected %v to be equal to %v", a, a)
}
