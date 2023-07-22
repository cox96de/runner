package kube

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/portforward"
)

func TestRunner_GetExecutor(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		r := &Runner{
			executorPortMap: map[string]int32{"test": 1},
			pod: &corev1.Pod{
				Status: corev1.PodStatus{PodIP: "192.168.31.2"},
			},
		}
		executor, err := r.GetExecutor(context.Background(), "test")
		assert.NilError(t, err)
		assert.Assert(t, executor != nil)
	})
	t.Run("no_port", func(t *testing.T) {
		r := &Runner{
			executorPortMap: map[string]int32{},
		}
		_, err := r.GetExecutor(context.Background(), "test")
		assert.ErrorContains(t, err, "not found")
	})
	t.Run("port_forward", func(t *testing.T) {
		r := &Runner{
			portForwarder:      &portforward.PortForwarder{},
			portForwardPortMap: map[string]int32{"test": 1},
		}
		executor, err := r.GetExecutor(context.Background(), "test")
		assert.NilError(t, err)
		assert.Assert(t, executor != nil)
	})
}
