package starlark

import (
	"bytes"
	"os"
	"runtime"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

type buf struct {
	bytes.Buffer
}

func (b *buf) Close() error {
	return nil
}

func TestNewCommand(t *testing.T) {
	t.Run("complex_script", func(t *testing.T) {
		buf := &buf{}
		script := `
if platform_system() == "Linux":
   subprocess.run(["ls", "-alh"])
if platform_system() == "Darwin":
   subprocess.run(["ls", "-alh"])
if platform_system() == "Windows":
   subprocess.run(["powershell","-c","Write-Host execute.go"])
`
		starlark, err := NewCommand(script, buf, buf, "", os.Environ(), "")
		assert.NilError(t, err)
		err = starlark.Start()
		assert.NilError(t, err)
		err = <-starlark.Wait()
		assert.NilError(t, err)
		assert.Assert(t, strings.Contains(buf.String(), "execute.go"), buf.String())
	})
	t.Run("print", func(t *testing.T) {
		buf := &buf{}
		starlark, err := NewCommand(`print("hello")`, buf, buf, "", os.Environ(), "")
		assert.NilError(t, err)
		err = starlark.Start()
		assert.NilError(t, err)
		err = <-starlark.Wait()
		assert.NilError(t, err)
		assert.Assert(t, buf.String() == "hello")
	})
	t.Run("subprocess_run", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("skipping on windows")
		}
		buf := &buf{}
		starlark, err := NewCommand(`subprocess.run(["ls","-alh"])`, buf, buf, "", os.Environ(), "")
		assert.NilError(t, err)
		err = starlark.Start()
		assert.NilError(t, err)
		err = <-starlark.Wait()
		assert.NilError(t, err)
	})
	t.Run("platform_system", func(t *testing.T) {
		buf := &buf{}
		starlark, err := NewCommand(`print(platform.system())`, buf, buf, "", os.Environ(), "")
		assert.NilError(t, err)
		err = starlark.Start()
		assert.NilError(t, err)
		err = <-starlark.Wait()
		assert.NilError(t, err)
		var expected string
		switch runtime.GOOS {
		case "linux":
			expected = "Linux"
		case "darwin":
			expected = "Darwin"
		case "windows":
			expected = "Windows"
		}
		assert.Assert(t, buf.String() == expected, buf.String())
	})
	t.Run("test", func(t *testing.T) {
		buf := &buf{}
		starlark, err := NewCommand(`print(os.environment)`, buf, buf, "", os.Environ(), "")
		assert.NilError(t, err)
		err = starlark.Start()
		assert.NilError(t, err)
		err = <-starlark.Wait()
		assert.NilError(t, err)
		t.Logf("%s", buf.String())
	})
}
