package handler

import (
	"bytes"
	"strings"
)

const (
	// Taken from GitLab Runner.
	psHeader       = "& {\r\n"
	psOptionScript = "$ErrorActionPreference=\"Stop\";\r\n"
	psCheckError   = "if(!$?) { Exit &{if($LASTEXITCODE) {$LASTEXITCODE} else {1}} }\r\n"
	psFooter       = "}\r\n\r\n"
)

// compileWindowsScript compiles commands into a PowerShell script.
func compileWindowsScript(commands []string) string {
	buf := bytes.NewBufferString(psHeader + psOptionScript)
	for _, command := range commands {
		prompt := command
		// Taken from: http://www.robvanderwoude.com/escapechars.php
		prompt = strings.ReplaceAll(prompt, "`", "``")
		prompt = strings.ReplaceAll(prompt, "\a", "`a")
		prompt = strings.ReplaceAll(prompt, "\b", "`b")
		prompt = strings.ReplaceAll(prompt, "\f", "^f")
		prompt = strings.ReplaceAll(prompt, "\r", "`r")
		prompt = strings.ReplaceAll(prompt, "\n", "`n")
		prompt = strings.ReplaceAll(prompt, "\t", "^t")
		prompt = strings.ReplaceAll(prompt, "\v", "^v")
		prompt = strings.ReplaceAll(prompt, "#", "`#")
		prompt = strings.ReplaceAll(prompt, "'", "`'")
		prompt = strings.ReplaceAll(prompt, "\"", "`\"")
		prompt = strings.ReplaceAll(prompt, "$", "`$")
		prompt = strings.ReplaceAll(prompt, "``e", "`e")
		buf.WriteString(`"+ ` + prompt + `"` + "\r\n")
		buf.WriteString(command + "\r\n")
		buf.WriteString(psCheckError)
	}
	buf.WriteString(psFooter)
	return buf.String()
}
