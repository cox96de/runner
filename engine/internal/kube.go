package internal

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	watchtool "k8s.io/client-go/tools/watch"
)

func WaitPodReady(ctx context.Context, client kubernetes.Interface, p *corev1.Pod) (*corev1.Pod, error) {
	watcher, err := watchtool.NewRetryWatcher(p.ResourceVersion, &cache.ListWatch{
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return client.CoreV1().Pods(p.Namespace).Watch(ctx, metav1.ListOptions{
				FieldSelector: "metadata.name=" + p.Name,
			})
		},
	})
	if err != nil {
		return p, errors.WithMessage(err, "failed to create retry watcher")
	}
	defer watcher.Stop()
	for {
		select {
		case <-ctx.Done():
			return p, errors.WithStack(ctx.Err())
		case e := <-watcher.ResultChan():
			switch e.Type {
			case watch.Modified:
				pod := e.Object.(*corev1.Pod)
				switch pod.Status.Phase {
				case corev1.PodPending:
					if err := checkContainsStatus(pod); err != nil {
						return pod, err
					}
					continue
				case corev1.PodRunning:
					return pod, nil
				case corev1.PodSucceeded, corev1.PodFailed:
					return pod, errors.Errorf("pod is %s", pod.Status.Phase)
				default:
					return pod, errors.Errorf("unknown pod phase %s", pod.Status.Phase)
				}
			}
		}
	}
}

func checkContainsStatus(pod *corev1.Pod) error {
	errs := &multierror.Error{}
	for _, c := range pod.Status.ContainerStatuses {
		if c.State.Waiting != nil {
			switch c.State.Waiting.Reason {
			case "ImagePullBackOff", "ErrImagePull":
				errs = multierror.Append(errs, errors.Errorf("container %s is %s, message: %s",
					c.Name, c.State.Waiting.Reason, c.State.Waiting.Message))
			}
		}
	}
	return errs.ErrorOrNil()
}
