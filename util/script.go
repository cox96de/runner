package util

import (
	"bytes"
	"fmt"
)

// CompileUnixScript compiles unix commands into a single shell script which can be executed by shell.
func CompileUnixScript(commands []string) string {
	return compileUnixScript(commands)
}

func compileUnixScript(commands []string) string {
	buf := bytes.NewBufferString(shellHeader)
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
		// Convert character to octal format to avoid problems with special characters.
		for _, c := range []byte(string(r)) {
			b.WriteString(fmt.Sprintf(`\%03o`, c))
		}
	}
	return b.String()
}

const shellHeader = `
set -e

`

const traceScript = `
printf '+ %s\n'

%s
`
