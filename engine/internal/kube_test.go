package internal

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/watch"
	k8stest "k8s.io/client-go/testing"

	"gotest.tools/v3/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestWaitPodReady(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()
		fakeWatcher := watch.NewFake()
		clientset.PrependWatchReactor("pods", k8stest.DefaultWatchReactor(fakeWatcher, nil))
		pod := corev1.Pod{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				ResourceVersion: "1",
				Name:            t.Name(),
			},
			Spec:   corev1.PodSpec{},
			Status: corev1.PodStatus{},
		}
		go func() {
			pod1 := pod
			pod1.Status.Phase = corev1.PodPending
			fakeWatcher.Modify(&pod1)
			pod2 := pod
			pod2.Status.Phase = corev1.PodRunning
			fakeWatcher.Modify(&pod2)
			fakeWatcher.Stop()
		}()
		readPod, err := WaitPodReady(context.Background(), clientset, &pod)
		assert.NilError(t, err)
		assert.Equal(t, readPod.Status.Phase, corev1.PodRunning)
	})
	t.Run("to_completed", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()
		fakeWatcher := watch.NewFake()
		clientset.PrependWatchReactor("pods", k8stest.DefaultWatchReactor(fakeWatcher, nil))
		pod := corev1.Pod{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				ResourceVersion: "1",
				Name:            t.Name(),
			},
			Spec: corev1.PodSpec{},
			Status: corev1.PodStatus{
				Phase: corev1.PodPending,
			},
		}
		go func() {
			pod1 := pod
			pod.Status.Phase = corev1.PodPending
			fakeWatcher.Modify(&pod1)
			pod2 := pod
			pod2.Status.Phase = corev1.PodFailed
			fakeWatcher.Modify(&pod2)
			fakeWatcher.Stop()
		}()

		readPod, err := WaitPodReady(context.Background(), clientset, &pod)
		assert.ErrorContains(t, err, "Failed")
		assert.Equal(t, readPod.Status.Phase, corev1.PodFailed)
	})
	t.Run("bad_image", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()
		fakeWatcher := watch.NewFake()
		clientset.PrependWatchReactor("pods", k8stest.DefaultWatchReactor(fakeWatcher, nil))
		pod := corev1.Pod{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				ResourceVersion: "1",
				Name:            t.Name(),
			},
			Spec: corev1.PodSpec{},
			Status: corev1.PodStatus{
				Phase: corev1.PodPending,
			},
		}
		go func() {
			pod1 := pod
			pod1.Status.Phase = corev1.PodPending
			pod1.Status.ContainerStatuses = []corev1.ContainerStatus{
				{
					State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{
						Reason: "ErrImagePull",
					}},
				},
			}
			fakeWatcher.Modify(&pod1)
		}()

		readPod, err := WaitPodReady(context.Background(), clientset, &pod)
		assert.ErrorContains(t, err, "ErrImagePull")
		assert.Assert(t, readPod != nil)
	})
}
