package kube

import (
	"context"

	"github.com/cox96de/runner/engine"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Runner struct {
	client    kubernetes.Interface
	pod       *v1.Pod
	namespace string
}

func (r *Runner) Start(ctx context.Context) error {
	createdPod, err := r.client.CoreV1().Pods(r.namespace).Create(ctx, r.pod, metav1.CreateOptions{})
	if err != nil {
		return errors.WithStack(err)
	}
	r.pod = createdPod
	return nil
}

func (r *Runner) GetExecutor(ctx context.Context) (engine.Executor, error) {
	// TODO: implement
	return nil, errors.Errorf("not implemented")
}

func (r *Runner) Stop(ctx context.Context) error {
	err := r.client.CoreV1().Pods(r.namespace).Delete(ctx, r.pod.Name, metav1.DeleteOptions{})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
