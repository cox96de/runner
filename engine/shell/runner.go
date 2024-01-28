package shell

import (
	"context"
	"net"
	"net/http"

	"github.com/cox96de/runner/app/executor/handler"
	"github.com/cox96de/runner/engine"
	"github.com/cox96de/runner/internal/executor"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Runner struct {
	executorHandler *handler.Handler
	s               *http.Server
	addr            string
}

func NewRunner() *Runner {
	return &Runner{
		executorHandler: handler.NewHandler(),
	}
}

func (r *Runner) Start(_ context.Context) error {
	g := gin.New()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return errors.WithMessage(err, "failed to listen on a random port")
	}
	r.addr = listener.Addr().String()
	log.Infof("executor is serving on %s", r.addr)
	r.executorHandler.RegisterRoutes(g)
	r.s = &http.Server{
		Handler: g.Handler(),
	}
	go func() {
		err := r.s.Serve(listener)
		if err != nil {
			log.Errorf("executor handler is stopped: %+v", err)
		}
		log.Infof("server is stopped")
	}()
	return nil
}

func (r *Runner) GetExecutor(_ context.Context, _ string) (engine.Executor, error) {
	return executor.NewClient("http://" + r.addr), nil
}

func (r *Runner) Stop(ctx context.Context) error {
	return r.s.Shutdown(ctx)
}
