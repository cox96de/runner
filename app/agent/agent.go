package agent

import (
	"context"
	"time"

	"github.com/cox96de/runner/log"

	"github.com/cox96de/runner/api"

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
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		job, err := a.poll(ctx, time.Second)
		if err != nil {
			log.Errorf("failed to poll job: %v", err)
			continue
		}
		logger := log.ExtractLogger(ctx).WithFields(log.Fields{"job": job.ID, "job_execution": job.Executions[0].ID})
		logger.Infof("got job")
		err = NewExecution(a.engine, job, a.client).
			Execute(log.WithLogger(ctx, logger))
		if err != nil {
			logger.Errorf("failed to execute job: %v", err)
		}
	}
}
