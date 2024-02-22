package agent

import (
	"context"
	"time"

	"github.com/cox96de/runner/api"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"

	"github.com/cox96de/runner/engine"
)

type Agent struct {
	engine engine.Engine
	client api.ServerClient
}

func NewAgent(engine engine.Engine, client api.ServerClient) *Agent {
	return &Agent{engine: engine, client: client}
}

func (a *Agent) poll(ctx context.Context, interval time.Duration) (*api.Job, error) {
	for {
		requestJobResponse, err := a.client.RequestJob(ctx, &api.RequestJobRequest{})
		if err != nil {
			return nil, err
		}
		job := requestJobResponse.Job
		if job != nil {
			return job, nil
		}
		if job == nil {
			time.Sleep(interval)
		}
	}
}

func (a *Agent) Run(ctx context.Context) error {
	for {
		job, err := a.poll(ctx, time.Second)
		if err != nil {
			return errors.WithMessage(err, "failed to poll job")
		}
		log.Infof("got job: %d, execution: %d", job.ID, job.Executions[0].ID)
		err = NewExecution(a.engine, job, a.client).
			Execute(ctx)
		if err != nil {
			log.Errorf("failed to execute job: %v", err)
		}
	}
}
