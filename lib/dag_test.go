package lib

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
)

type n struct {
	id  string
	dep []string
}

func (nn *n) Depends() []string {
	return nn.dep
}

func (nn *n) ID() string {
	return nn.id
}

func TestNewDAG(t *testing.T) {
	type args struct {
		rawDAG map[string][]string
	}
	type testCase struct {
		name            string
		args            args
		expectDeepPosts map[string][]string
		expectDeepPres  map[string][]string
		expectError     bool
	}
	tests := []testCase{
		{
			name: "no_dep",
			args: args{
				rawDAG: map[string][]string{
					"a": {},
					"b": {},
				},
			},
			expectDeepPosts: map[string][]string{
				"a": {},
				"b": {},
			},
			expectDeepPres: map[string][]string{
				"a": {},
				"b": {},
			},
		},
		{
			name: "dep",
			args: args{
				rawDAG: map[string][]string{
					"a": {},
					"b": {"a"},
				},
			},
			expectDeepPosts: map[string][]string{
				"a": {"b"},
				"b": {},
			},
			expectDeepPres: map[string][]string{
				"a": {},
				"b": {"a"},
			},
		},
		{
			name: "dep",
			args: args{
				rawDAG: map[string][]string{
					"a": {},
					"b": {"a"},
					"c": {"b"},
				},
			},
			expectDeepPosts: map[string][]string{
				"a": {"b", "c"},
				"b": {"c"},
				"c": {},
			},
			expectDeepPres: map[string][]string{
				"a": {},
				"b": {"a"},
				"c": {"a", "b"},
			},
		},
		{
			name: "dep",
			args: args{
				rawDAG: map[string][]string{
					"a": {},
					"b": {"a"},
					"c": {},
				},
			},
			expectDeepPosts: map[string][]string{
				"a": {"b"},
				"b": {},
				"c": {},
			},
			expectDeepPres: map[string][]string{
				"a": {},
				"b": {"a"},
				"c": {},
			},
		},
		{
			name: "loop",
			args: args{
				rawDAG: map[string][]string{
					"a": {"c"},
					"b": {"a"},
					"c": {"b"},
				},
			},
			expectError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()
			testDAG(t, tt.args.rawDAG, tt.expectDeepPosts, tt.expectDeepPres, tt.expectError)
		})
	}
}

func testDAG(t *testing.T, rawDAG map[string][]string, expectDeepPosts map[string][]string, expectDeepPres map[string][]string,
	expectError bool,
) {
	t.Helper()
	nodes := make([]*n, 0, len(rawDAG))
	for node, deps := range rawDAG {
		nodes = append(nodes, &n{id: node, dep: deps})
	}
	dag, err := NewDAG(nodes...)
	if expectError {
		assert.Assert(t, err != nil)
		return
	}
	for _, n2 := range nodes {
		deepPre, err := dag.DeepPre(n2.ID())
		assert.NilError(t, err)
		deepPreIDs := lo.Map(deepPre, func(item *n, index int) string {
			return item.id
		})
		assert.DeepEqual(t, deepPreIDs, expectDeepPres[n2.ID()], cmpopts.SortSlices(func(a, b string) bool {
			return strings.Compare(a, b) > 0
		}))

		deepPost, err := dag.DeepPost(n2.ID())
		assert.NilError(t, err)
		deepPostIDs := lo.Map(deepPost, func(item *n, index int) string {
			return item.id
		})
		assert.DeepEqual(t, deepPostIDs, expectDeepPosts[n2.ID()], cmpopts.SortSlices(func(a, b string) bool {
			return strings.Compare(a, b) > 0
		}))
	}
}
