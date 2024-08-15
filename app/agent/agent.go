package agent

import (
	"context"
	"time"

	"github.com/cox96de/runner/util"

	"github.com/cox96de/runner/log"

	"github.com/cox96de/runner/api"

	"github.com/cox96de/runner/engine"
)

type Agent struct {
	engine engine.Engine
	client api.ServerClient
	label  string
}

func NewAgent(engine engine.Engine, client api.ServerClient, label string) *Agent {
	return &Agent{engine: engine, client: client, label: label}
}

func (a *Agent) poll(ctx context.Context, interval time.Duration) (*api.Job, error) {
	for {
		requestJobResponse, err := a.client.RequestJob(ctx, &api.RequestJobRequest{
			Label: a.label,
		})
		if err != nil {
			log.Errorf("failed to request job: %v", err)
			if err := util.Wait(ctx, interval); err != nil {
				return nil, err
			}
			continue
		}
		job := requestJobResponse.Job
		if job != nil {
			return job, nil
		}
		if job == nil {
			if err := util.Wait(ctx, interval); err != nil {
				return nil, err
			}
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
		interval := time.Second
		job, err := a.poll(ctx, interval)
		if err != nil {
			return err
		}
		logger := log.ExtractLogger(ctx).WithFields(log.Fields{"job": job.ID, "job_execution": job.Execution.ID})
		logger.Infof("got job")
		err = NewExecution(a.engine, job, a.client).
			Execute(log.WithLogger(ctx, logger))
		if err != nil {
			logger.Errorf("failed to execute job: %v", err)
		}
	}
}
