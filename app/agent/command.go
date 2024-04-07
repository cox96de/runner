package agent

import (
	"bytes"
	"fmt"
	"strings"
)

const shellOptionScript = `
set -e

`

const traceScript = `
printf '+ %s\n'

%s
`

// compileUnixScript compiles unix commands into a single command which can be executed by shell.
func compileUnixScript(commands []string) string {
	buf := bytes.NewBufferString(shellOptionScript)
	for _, command := range commands {
		escaped := encodeCommandLine(command)
		buf.WriteString(fmt.Sprintf(
			traceScript,
			escaped,
			command,
		))
	}
	return buf.String()
}

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

func skipEncode(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || (r == ' ')
}

func encodeCommandLine(l string) string {
	b := &bytes.Buffer{}
	for _, r := range l {
		if skipEncode(r) {
			b.WriteRune(r)
			continue
		}
		for _, c := range []byte(string(r)) {
			b.WriteString(fmt.Sprintf(`\%03o`, c))
		}
	}
	return b.String()
}

func getUnixCommands() []string {
	return []string{"/bin/sh", "-c", "printf '%s' \"$RUNNER_SCRIPT\" | /bin/sh"}
}

func getWindowsCommands() []string {
	return []string{"powershell.exe", "-NoProfile", "-c", "echo $env:RUNNER_SCRIPT | powershell.exe -NoProfile " +
		"-NoLogo -InputFormat text -OutputFormat text -ExecutionPolicy Bypass -NonInteractive -Command -"}
}
