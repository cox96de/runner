package agent

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
			args: args{commands: []string{"go build -o server"}},
			want: `
set -e


printf '+ go build \055o server\n'

go build -o server
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compileUnixScript(tt.args.commands); got != tt.want {
				t.Errorf("compileUnixScript() = %v, want %v", got, tt.want)
			}
		})
	}
}
