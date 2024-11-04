package agent

import (
	"context"
	"sync"
	"time"

	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/engine"
	"github.com/cox96de/runner/log"
	"github.com/cox96de/runner/util"
	"github.com/pkg/errors"
)

var StopError = util.StringError("agent is closed")

type Agent struct {
	engine engine.Engine
	client api.ServerClient
	label  string
	stop   chan struct{}
}

func NewAgent(engine engine.Engine, client api.ServerClient, label string) *Agent {
	return &Agent{engine: engine, client: client, label: label, stop: make(chan struct{})}
}

func (a *Agent) poll(ctx context.Context, ticker chan struct{}) (*api.Job, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-a.stop:
			return nil, StopError
		case <-ticker:
			requestJobResponse, err := a.client.RequestJob(ctx, &api.RequestJobRequest{
				Label: a.label,
			})
			if err != nil {
				log.Errorf("failed to request job: %v", err)
				continue
			}
			if err != nil {
				log.Errorf("failed to request job: %v", err)
				continue
			}
			job := requestJobResponse.Job
			if job != nil {
				select {
				// Accept new job immediately, no need to wait for a period.
				case ticker <- struct{}{}:
				default:

				}
				return job, nil
			}
		}
	}
}

func (a *Agent) Run(ctx context.Context, concurrency int, fetchInterval time.Duration) error {
	wg := &sync.WaitGroup{}
	ticker := make(chan struct{})
	go func() {
		t := time.NewTicker(fetchInterval)
		for {
			<-t.C
			select {
			case ticker <- struct{}{}:
			case <-a.stop:
				return
			}
		}
	}()
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				job, err := a.poll(ctx, ticker)
				if err != nil {
					if errors.Is(err, StopError) {
						return
					}
					log.Warningf("failed to poll job: %v", err)
					time.Sleep(time.Second * 5)
					continue
				}
				logger := log.ExtractLogger(ctx).WithFields(log.Fields{"job": job.ID, "job_execution": job.Execution.ID})
				logger.Infof("got job")
				err = NewExecution(a.engine, job, a.client).
					Execute(log.WithLogger(ctx, logger))
				if err != nil {
					logger.Errorf("failed to execute job: %v", err)
				}
			}
		}()

	}
	<-a.stop
	log.Infof("exiting, wait for all jobs to finish")
	wg.Wait()
	return nil
}

// GracefulShutdown stops to accept new jobs and await all accepted jobs to be completed and to exit.
func (a *Agent) GracefulShutdown() {
	close(a.stop)
}
