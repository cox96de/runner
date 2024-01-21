package shell

import (
	"context"
	"net"

	"github.com/cox96de/runner/app/executor/executorpb"
	"github.com/cox96de/runner/app/executor/handler"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Runner struct {
	executorHandler *handler.Handler
	s               *grpc.Server
	addr            string
}

func NewRunner() *Runner {
	return &Runner{
		executorHandler: handler.NewHandler(),
	}
}

func (r *Runner) Start(_ context.Context) error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return errors.WithMessage(err, "failed to listen on a random port")
	}
	r.addr = listener.Addr().String()
	log.Infof("executor is serving on %s", r.addr)
	r.s = grpc.NewServer()
	executorpb.RegisterExecutorServer(r.s, r.executorHandler)
	go func() {
		err := r.s.Serve(listener)
		if err != nil {
			log.Errorf("executor handler is stopped: %+v", err)
		}
		log.Infof("server is stopped")
	}()
	return nil
}

func (r *Runner) GetExecutor(_ context.Context, _ string) (executorpb.ExecutorClient, error) {
	conn, err := grpc.Dial(r.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.WithMessage(err, "failed to connect to executor")
	}
	return executorpb.NewExecutorClient(conn), nil
}

func (r *Runner) Stop(ctx context.Context) error {
	r.s.Stop()
	return nil
}
