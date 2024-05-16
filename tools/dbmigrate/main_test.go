package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_migrate(t *testing.T) {
	// TODO: collect the output of migrate.
	err := migrate("sqlite", "file::memory:", "")
	assert.NoError(t, err)
}
