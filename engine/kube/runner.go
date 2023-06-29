package kube

import (
	"context"
	"fmt"
	"net"

	"github.com/cox96de/runner/engine"
	"github.com/cox96de/runner/internal/executor"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	watchtool "k8s.io/client-go/tools/watch"
)

type Runner struct {
	client          kubernetes.Interface
	pod             *v1.Pod
	executorPortMap map[string]int32
	namespace       string
}

func (r *Runner) Start(ctx context.Context) error {
	createdPod, err := r.client.CoreV1().Pods(r.namespace).Create(ctx, r.pod, metav1.CreateOptions{})
	if err != nil {
		return errors.WithStack(err)
	}
	r.pod = createdPod
	err = r.waitPodReady(ctx)
	return err
}

func (r *Runner) waitPodReady(ctx context.Context) error {
	watcher, err := watchtool.NewRetryWatcher(r.pod.ResourceVersion, &cache.ListWatch{
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return r.client.CoreV1().Pods(r.namespace).Watch(ctx, metav1.ListOptions{
				FieldSelector: "metadata.name=" + r.pod.Name,
			})
		},
	})
	if err != nil {
		return errors.WithStack(err)
	}
	for {
		select {
		case <-ctx.Done():
			return errors.WithStack(ctx.Err())
		case e := <-watcher.ResultChan():
			switch e.Type {
			case watch.Modified:
				pod := e.Object.(*v1.Pod)
				switch pod.Status.Phase {
				case corev1.PodPending:
					continue
				case corev1.PodRunning:
					r.pod = pod
					return nil
				case corev1.PodSucceeded:
					return errors.New("pod is succeeded")
				case corev1.PodFailed:
					return errors.New("pod is failed")
				default:
					return errors.Errorf("unknown pod phase %s", pod.Status.Phase)
				}
			}
		}
	}
}

func (r *Runner) GetExecutor(ctx context.Context, name string) (engine.Executor, error) {
	port, ok := r.executorPortMap[name]
	if !ok {
		return nil, errors.Errorf("the runner container %s not found", name)
	}
	client := executor.NewClient(fmt.Sprintf("http://%s", net.JoinHostPort(r.pod.Status.PodIP,
		fmt.Sprintf("%d", port))))
	return client, nil
}

func (r *Runner) Stop(ctx context.Context) error {
	err := r.client.CoreV1().Pods(r.namespace).Delete(ctx, r.pod.Name, metav1.DeleteOptions{})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
