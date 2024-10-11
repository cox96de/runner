package vm

import (
	"context"
	"testing"

	"github.com/cox96de/runner/app/executor/executorpb"
	"github.com/cox96de/runner/app/executor/executorpb/mock"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	k8stest "k8s.io/client-go/testing"
)

func TestRunner_Start(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	fakeWatcher := watch.NewFake()
	clientset.PrependWatchReactor("pods", k8stest.DefaultWatchReactor(fakeWatcher, nil))
	pod := corev1.Pod{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "vm",
			ResourceVersion: "1",
			Namespace:       "namespace",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodPending,
		},
	}
	go func() {
		pod := pod
		pod.Status.PodIP = "fakeip"
		pod.Status.Phase = corev1.PodRunning
		fakeWatcher.Modify(&pod)
		fakeWatcher.Stop()
	}()
	r := &Runner{
		client:    clientset,
		pod:       &pod,
		port:      50051,
		namespace: "namespace",
		executorDialer: func(ctx context.Context, address string) (executorpb.ExecutorClient, error) {
			assert.Equal(t, address, "fakeip:50051")
			client := mock.NewMockExecutorClient(gomock.NewController(t))
			client.EXPECT().Ping(gomock.Any(), gomock.Any()).Return(&executorpb.PingResponse{}, nil).AnyTimes()
			return client, nil
		},
	}
	err := r.Start(context.Background())
	assert.NilError(t, err)
}
