package testtool

import (
	"os/exec"
	"strings"

	"github.com/cockroachdb/errors"
)

// GetRepositoryRoot returns the root path of the repository.zazz
func GetRepositoryRoot() (string, error) {
	resultByte, err := exec.Command("git", "rev-parse", "--show-toplevel").CombinedOutput()
	if err != nil {
		return "", errors.WithMessagef(err, "output: %s", string(resultByte))
	}
	path := string(resultByte)
	path = strings.TrimRight(path, "\n")
	return path, nil
}
