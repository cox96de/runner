package kube

import (
	"context"
	"fmt"
	"net"

	"github.com/cox96de/runner/app/executor/executorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"k8s.io/client-go/tools/portforward"

	"github.com/hashicorp/go-multierror"

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
	defaultContainer   string
	client             kubernetes.Interface
	pod                *v1.Pod
	executorPortMap    map[string]int32
	portForwardPortMap map[string]int32
	namespace          string
	portForwarder      *portforward.PortForwarder
	portForwardStop    chan struct{}
}

func (r *Runner) Start(ctx context.Context) (startErr error) {
	createdPod, err := r.client.CoreV1().Pods(r.namespace).Create(ctx, r.pod, metav1.CreateOptions{})
	if err != nil {
		return errors.WithStack(err)
	}
	r.pod = createdPod
	err = r.waitPodReady(ctx)
	if err != nil {
		// Clean up the created kube resources if about to fail to avoid resource leak.
		if cleanErr := r.clean(ctx); cleanErr != nil {
			startErr = multierror.Append(startErr, cleanErr)
		}
	}
	if r.portForwarder != nil {
		err := r.waitPortForwarderReady(ctx)
		if err != nil {
			// Clean up the created kube resources if about to fail to avoid resource leak.
			if cleanErr := r.clean(ctx); cleanErr != nil {
				startErr = multierror.Append(startErr, cleanErr)
			}
		}
	}
	return startErr
}

func (r *Runner) waitPortForwarderReady(ctx context.Context) error {
	go func() {
		_ = r.portForwarder.ForwardPorts() // TODO: handle error
	}()
	r.portForwardPortMap = make(map[string]int32)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-r.portForwarder.Ready:
		ports, err := r.portForwarder.GetPorts()
		if err != nil {
			return errors.WithStack(err)
		}
		for _, port := range ports {
			for k, p := range r.executorPortMap {
				if int32(port.Remote) == p {
					r.portForwardPortMap[k] = int32(port.Local)
					break
				}
			}
		}
		return nil
	}
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
	defer watcher.Stop()
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
				case corev1.PodSucceeded, corev1.PodFailed:
					return errors.Errorf("pod is %s", pod.Status.Phase)
				default:
					return errors.Errorf("unknown pod phase %s", pod.Status.Phase)
				}
			}
		}
	}
}

func (r *Runner) GetExecutor(ctx context.Context) (executorpb.ExecutorClient, error) {
	return r.GetContainerExecutor(ctx, r.defaultContainer)
}

func (r *Runner) GetContainerExecutor(ctx context.Context, containerName string) (executorpb.ExecutorClient, error) {
	if r.portForwarder != nil {
		return r.getExecutorFromPortForward(containerName)
	}
	port, ok := r.executorPortMap[containerName]
	if !ok {
		return nil, errors.Errorf("the runner container %s not found", containerName)
	}
	addr := net.JoinHostPort(r.pod.Status.PodIP,
		fmt.Sprintf("%d", port))
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.WithMessage(err, "failed to connect to executor")
	}
	client := executorpb.NewExecutorClient(conn)
	return client, nil
}

func (r *Runner) getExecutorFromPortForward(name string) (executorpb.ExecutorClient, error) {
	port, ok := r.portForwardPortMap[name]
	if !ok {
		return nil, errors.Errorf("the runner container %s not found", name)
	}
	addr := net.JoinHostPort("127.0.0.1",
		fmt.Sprintf("%d", port))
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.WithMessage(err, "failed to connect to executor")
	}
	client := executorpb.NewExecutorClient(conn)
	return client, nil
}

func (r *Runner) Stop(ctx context.Context) error {
	return r.clean(ctx)
}

func (r *Runner) clean(ctx context.Context) error {
	var multierr *multierror.Error
	if r.pod != nil {
		err := r.client.CoreV1().Pods(r.namespace).Delete(ctx, r.pod.Name, metav1.DeleteOptions{})
		if err != nil {
			multierr = multierror.Append(multierr, err)
		}
	}
	if r.portForwarder != nil {
		close(r.portForwardStop)
	}
	return multierr.ErrorOrNil()
}
