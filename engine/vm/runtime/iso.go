//go:build linux

package main

import (
	"os/exec"

	"github.com/pkg/errors"
)

// genISO generates an iso file.
func genISO(label string, output string, inputs ...string) error {
	combinedOutput, err := exec.Command("genisoimage",
		append([]string{"-o", output, "-V", label, "-r", "-J", "-quiet"}, inputs...)...).CombinedOutput()
	if err != nil {
		return errors.WithMessagef(err, "failed to generate iso: %s", string(combinedOutput))
	}
	return nil
}
