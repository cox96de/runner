package handler

import (
	"testing"

	"gotest.tools/v3/assert"
)

func Test_compileWindowsScript(t *testing.T) {
	script := compileWindowsScript([]string{"Write-Host \"Hello World\""})
	assert.DeepEqual(t, script, "& {\r\n$ErrorActionPreference=\"Stop\";\r\n\"+ Write-Host `\"Hello World`\"\"\r\nWrite-Host \"Hello World\"\r\nif(!$?) { Exit &{if($LASTEXITCODE) {$LASTEXITCODE} else {1}} }\r\n}\r\n\r\n")
}
