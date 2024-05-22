package dag

import (
	"context"
	"runtime"
	"runtime/debug"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
)

type testRunner struct {
	duration  time.Duration
	err       error
	startedAt *time.Time
	completed *time.Time
}

var now = func() time.Time {
	return time.Now()
}

func (t *testRunner) Run(ctx context.Context) error {
	t.startedAt = lo.ToPtr(now())
	defer func() {
		t.completed = lo.ToPtr(now())
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(t.duration):
		return t.err
	}
}

func TestNewRunner(t *testing.T) {
	t.Run("no error run", func(t *testing.T) {
		testNormalDAG(t, map[string][]string{
			"A": {"B", "C"},
			"B": {},
			"C": {},
		})
		testNormalDAG(t, map[string][]string{
			"A": {},
			"B": {"A"},
			"C": {"A", "B"},
			"D": {},
			"E": {"C", "D"},
			"F": {"E"},
		})
		testNormalDAG(t, map[string][]string{
			"A": nil,
			"B": {"A"},
			"C": {"B"},
			"D": {"C"},
		})
		testNormalDAG(t, map[string][]string{
			"A": {"C"},
			"B": {"A"},
			"C": {"D"},
			"D": {},
		})
		testNormalDAG(t, map[string][]string{
			"A": nil,
			"B": {"A"},
			"C": nil,
			"D": {"B"},
		})
		testNormalDAG(t, map[string][]string{
			"A": nil,
			"B": {"C"},
			"C": nil,
			"D": {"A"},
		})
		testNormalDAG(t, map[string][]string{
			"A": nil,
			"B": {"A", "C", "D"},
			"C": nil,
			"D": nil,
		})
		testNormalDAG(t, map[string][]string{
			"A": nil,
			"B": {"D"},
			"C": {"A", "B"},
			"D": nil,
		})
	})
	t.Run("step is error", func(t *testing.T) {
		testErrorDAG(t,
			map[string][]string{
				"A": {"B", "C"},
				"B": {},
				"C": {},
			},
			map[string]bool{"B": true},
			map[string]bool{"C": true},
		)

		testErrorDAG(t, map[string][]string{
			"A": {},
			"B": {"A"},
			"C": {"A", "B"},
			"D": {},
			"E": {"C", "D"},
			"F": {"E"},
		}, map[string]bool{"A": true},
			map[string]bool{
				"D": true,
			})
		testErrorDAG(t, map[string][]string{
			"A": nil,
			"B": {"A"},
			"C": {"B"},
			"D": {"C"},
		},
			map[string]bool{"C": true},
			map[string]bool{
				"A": true,
				"B": true,
			})
	})
	t.Run("panic test", func(t *testing.T) {
		runner := NewRunner()
		runner.AddVertex("A", func() error {
			panic("internal error")
		})
		runner.AddVertex("B", func() error {
			return nil
		})
		runner.AddEdge("A", "B")
		err := runner.Run()
		assert.ErrorContains(t, err, "recover from vertex")
	})
}

func testNormalDAG(t *testing.T, vertexes map[string][]string) {
	runners := make(map[string]*testRunner)
	expect := make(map[string]bool)
	for vertex := range vertexes {
		expect[vertex] = true
		if runtime.GOOS == "windows" {
			// Windows has a lower time resolution. It occurs in Github Actions.
			oldNow := now
			n := atomic.Int64{}
			n.Store(time.Now().Unix())
			now = func() time.Time {
				add := n.Add(1)
				return time.Unix(add, 0)
			}
			t.Cleanup(func() {
				now = oldNow
			})
		}
		runners[vertex] = &testRunner{
			duration: time.Millisecond * 10,
			err:      nil,
		}
	}
	testDAG(context.Background(), t, runners, vertexes, expect, "")
}

func testErrorDAG(t *testing.T, vertexes map[string][]string, errorNodes map[string]bool, expect map[string]bool) {
	runners := make(map[string]*testRunner)
	for vertex := range vertexes {
		var runnerErr error
		if errorNodes[vertex] {
			runnerErr = errors.Errorf("runner error for %s", vertex)
		}
		if runtime.GOOS == "windows" {
			// Windows has a lower time resolution. It occurs in Github Actions.
			oldNow := now
			n := atomic.Int64{}
			n.Store(time.Now().Unix())
			now = func() time.Time {
				add := n.Add(1)
				return time.Unix(add, 0)
			}
			t.Cleanup(func() {
				now = oldNow
			})
		}
		runners[vertex] = &testRunner{
			duration: time.Millisecond * 10,
			err:      runnerErr,
		}
	}
	testDAG(context.Background(), t, runners, vertexes, expect, "runner error for")
}

func testDAG(ctx context.Context, t *testing.T, runners map[string]*testRunner, depGraph map[string][]string, expect map[string]bool, expectErr string) {
	dag := NewRunner()

	for name, runner := range runners {
		runner := runner
		dag.AddVertex(name, func() error {
			return runner.Run(ctx)
		})
	}
	for runner, dependencies := range depGraph {
		for _, d := range dependencies {
			dag.AddEdge(d, runner)
		}
	}
	err := dag.Run()
	if len(expectErr) > 0 {
		assert.ErrorContains(t, err, expectErr)
	} else {
		assert.NilError(t, err)
	}
	validate(t, runners, depGraph, expect)
}

func validate(t *testing.T, runners map[string]*testRunner, depGraph map[string][]string, expect map[string]bool) {
	for vertexName, shouldExecute := range expect {
		vertexRunner := runners[vertexName]
		if shouldExecute {
			assert.Assert(t, vertexRunner.startedAt != nil, "startedAt is empty for vertex '%s', "+
				"stack: %s", vertexName, string(debug.Stack()))
			assert.Assert(t, vertexRunner.completed != nil, "completedAt is empty for vertex '%s', "+
				"stack: %s", vertexName, string(debug.Stack()))
			dependencies := depGraph[vertexName]
			for _, dependency := range dependencies {
				depRunner := runners[dependency]
				completed := depRunner.completed
				assert.Assert(t, vertexRunner.startedAt.After(*completed), "startedAt '%s' of vertex '%s' "+
					"should be after completedAt '%s' of vertex '%s' ", vertexName, vertexRunner.startedAt, dependency,
					depRunner.completed)
			}
		} else {
			assert.Assert(t, vertexRunner.startedAt == nil, "startedAt is not empty for %s", vertexName)
			assert.Assert(t, vertexRunner.completed == nil, "completedAt is not empty for %s", vertexName)
		}
	}
}
