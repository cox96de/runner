package vm

import (
	"reflect"
	"testing"
)

func Test_newCompiler(t *testing.T) {
	type args struct {
		executorImage string
		executorPath  string
		runtimeImage  string
	}
	tests := []struct {
		name string
		args args
		want *compiler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newCompiler(tt.args.executorImage, tt.args.executorPath, tt.args.runtimeImage); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newCompiler() = %v, want %v", got, tt.want)
			}
		})
	}
}
