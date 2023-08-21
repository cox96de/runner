package testtool

import (
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/env"
	"gotest.tools/v3/fs"
)

func TestGetRepositoryRoot(t *testing.T) {
	t.Run("git_repo", func(t *testing.T) {
		root, err := GetRepositoryRoot()
		assert.NilError(t, err)
		stat, err := os.Stat(filepath.Join(root, ".git"))
		assert.NilError(t, err)
		assert.DeepEqual(t, stat.Name(), ".git")
	})
	t.Run("non_git", func(t *testing.T) {
		dir := fs.NewDir(t, "gotest")
		env.ChangeWorkingDir(t, dir.Path())
		_, err := GetRepositoryRoot()
		assert.ErrorContains(t, err, "not a git repository")
	})
}
