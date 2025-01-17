package vm

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/cox96de/runner/engine"
	"github.com/cox96de/runner/log"

	"github.com/cox96de/runner/engine/internal"

	"github.com/cox96de/runner/app/executor/executorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/hashicorp/go-multierror"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Runner struct {
	client          kubernetes.Interface
	pod             *corev1.Pod
	portForwardPort int32
	port            int32
	namespace       string
	// executorDialer makes it easy to mock the dialer in tests.
	executorDialer func(ctx context.Context, address string) (executorpb.ExecutorClient, error)
	logProvider    engine.LogProvider
}

func (r *Runner) fetchLogs(ctx context.Context) {
	logger := log.ExtractLogger(ctx)
	logWriter := r.logProvider.CreateLogWriter(context.Background(), "_runtime")
	defer logWriter.Close()
	req := r.client.CoreV1().Pods(r.pod.Namespace).GetLogs(r.pod.Name, &corev1.PodLogOptions{
		Follow:    true,
		Container: r.pod.Spec.Containers[0].Name,
	})
	stream, err := req.Stream(context.Background())
	if err != nil {
		logger.Errorf("failed to start to fetch runtime's log: %v", err)
		return
	}
	defer stream.Close()
	buf := make([]byte, 2000)
	for {
		numBytes, err := stream.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Errorf("failed to read logs: %+v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		if numBytes == 0 {
			time.Sleep(1 * time.Second)
			continue
		}
		_, err = logWriter.Write(buf[:numBytes])
		if err != nil {
			logger.Errorf("failed to write logs: %+v", err)
		}
	}
}

func (r *Runner) Start(ctx context.Context) (startErr error) {
	createdPod, err := r.client.CoreV1().Pods(r.namespace).Create(ctx, r.pod, metav1.CreateOptions{})
	if err != nil {
		return errors.WithStack(err)
	}
	r.pod = createdPod
	r.pod, err = internal.WaitPodReady(ctx, r.client, r.pod)
	if err != nil {
		startErr = multierror.Append(startErr, err)
		// Clean up the created kube resources if about to fail to avoid resource leak.
		if cleanErr := r.clean(ctx); cleanErr != nil {
			startErr = multierror.Append(startErr, cleanErr)
		}
		return startErr
	}
	go r.fetchLogs(ctx)
	err = r.waitExecutorReady(ctx, time.Second, time.Minute*2)
	if err != nil {
		// Clean up the created kube resources if about to fail to avoid resource leak.
		if cleanErr := r.clean(ctx); cleanErr != nil {
			startErr = multierror.Append(startErr, cleanErr)
		}
	}
	return startErr
}

func (r *Runner) waitExecutorReady(ctx context.Context, interval, timeout time.Duration) error {
	executor, err := r.GetExecutor(ctx)
	if err != nil {
		return err
	}
	var (
		lastError error
		ticker    = time.NewTicker(interval)
	)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			_, lastError = executor.Ping(ctx, &executorpb.PingRequest{})
			if lastError == nil {
				return nil
			}
		case <-time.After(timeout):
			return errors.Errorf("timeout waiting for executor ready, last error: %+v", lastError)
		}
	}
}

func defaultExecutorDialer(addr string) (executorpb.ExecutorClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.WithMessage(err, "failed to connect to executor")
	}
	return executorpb.NewExecutorClient(conn), err
}

func (r *Runner) GetExecutor(ctx context.Context) (executorpb.ExecutorClient, error) {
	port := r.port
	addr := net.JoinHostPort(r.pod.Status.PodIP,
		fmt.Sprintf("%d", port))
	if r.executorDialer == nil {
		return defaultExecutorDialer(addr)
	}
	return r.executorDialer(ctx, addr)
}

func (r *Runner) Stop(ctx context.Context) error {
	return r.clean(ctx)
}

func (r *Runner) clean(ctx context.Context) error {
	multierr := &multierror.Error{}
	if r.pod != nil {
		err := r.client.CoreV1().Pods(r.namespace).Delete(ctx, r.pod.Name, metav1.DeleteOptions{})
		if err != nil {
			multierr = multierror.Append(multierr, err)
		}
	}
	return multierr.ErrorOrNil()
}
