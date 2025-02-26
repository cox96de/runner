package handler

import "testing"

func Test_compileUnixScript(t *testing.T) {
	type args struct {
		commands []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "simple",
			args: args{
				commands: []string{"echo hello"},
			},
			want: `
set -e


printf '+ echo hello\n'

echo hello
`,
		},
		{
			name: "env",
			args: args{
				commands: []string{"echo ${PATH}"},
			},
			want: `
set -e


printf '+ echo \044\173PATH\175\n'

echo ${PATH}
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compileUnixScript(tt.args.commands); got != tt.want {
				t.Errorf("CompileUnixScript() = %v, want %v", got, tt.want)
			}
		})
	}
}
