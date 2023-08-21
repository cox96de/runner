//go:build linux

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gotest.tools/v3/env"
	"gotest.tools/v3/fs"
)

func TestGenISO(t *testing.T) {
	dir := fs.NewDir(t, "test", fs.WithFile("test.txt", "test"))
	env.ChangeWorkingDir(t, dir.Path())
	err := genISO("test", "test.iso", "test.txt")
	if err != nil {
		t.Fatal(err)
	}
	assert.FileExists(t, "test.iso")
}
