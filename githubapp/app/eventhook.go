package app

import (
	"context"
	"encoding/json"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/gin-gonic/gin"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/githubapp/db"
	"github.com/cox96de/runner/log"
	"github.com/samber/lo"
)

func (h *App) refreshJob(ctx context.Context, job *db.Job) error {
	pipeline, err := h.db.GetPipelineByID(ctx, job.PipelineID)
	if err != nil {
		return errors.WithMessagef(err, "failed to get pipeline %d", job.PipelineID)
	}
	getJobExecutionResp, err := h.runnerClient.GetJobExecution(ctx, &api.GetJobExecutionRequest{
		JobExecutionID:    job.RunnerJobExecutionID,
		WithStepExecution: lo.ToPtr(true),
	})
	if err != nil {
		return err
	}
	var steps []*db.Step
	logger := log.ExtractLogger(ctx)
	if err = json.Unmarshal(job.Steps, &steps); err != nil {
		// Continue.
		logger.Errorf("failed to unmarashal steps: %+v", err)
	}
	logger.Infof("job execution status: %s", getJobExecutionResp.JobExecution.Status)
	jobExecution := getJobExecutionResp.JobExecution
	options, err := GenerateUpdateCheckRunOptions(h.baseURL, &RenderCheckRunOptions{
		RunnerJob: jobExecution,
		RenderJob: &RenderJob{
			UID:  job.UID,
			Name: job.Name,
			Steps: lo.Map(steps, func(step *db.Step, index int) *RenderStep {
				return &RenderStep{Name: step.Name}
			}),
		},
	})
	ghClient, err := h.ghClient.AppInstallClient(pipeline.AppInstallID)
	if err != nil {
		logger.Errorf("failed to get github client: %+v", err)
		return err
	}
	_, _, err = ghClient.Checks.UpdateCheckRun(ctx, pipeline.RepoOwner, pipeline.RepoName, job.CheckRunID, options)
	if err != nil {
		logger.Errorf("failed to update checkrun: %+v", err)
	}
	return nil
}

func (h *App) GetRunnerHandler(ctx context.Context) (gin.HandlerFunc, error) {
	proto, err := cloudevents.NewHTTP()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create cloud events protocal")
	}
	receiveHandler, err := cloudevents.NewHTTPReceiveHandler(ctx, proto, h.handleCloudEvents)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create http receive handler")
	}
	return gin.WrapH(receiveHandler), nil
}

func (h *App) handleCloudEvents(ctx context.Context, event cloudevents.Event) error {
	e := &api.Event{}
	err := event.DataAs(e)
	if err != nil {
		return errors.WithMessage(err, "failed to unmarshal cloud event data")
	}
	logger := log.ExtractLogger(ctx).WithField("event.id", event.ID())
	switch {
	case e.StepExecution != nil:
		return h.handleStepExecutionUpdate(log.WithLogger(ctx, logger), e.StepExecution)
	case e.JobExecution != nil:
		if e.JobExecution.Status == api.StatusQueued {
			// Runner dispatch the job to queue when the job is created and can be executed.
			// But the CI server might be not get the response of the job creation.
			// So cannot find the job in the database.
			return nil
		}
		return h.handleJobExecutionUpdate(log.WithLogger(ctx, logger), e.JobExecution)
	}
	return nil
}

func (h *App) handleStepExecutionUpdate(ctx context.Context, e *api.StepExecution) error {
	job, err := h.db.GetJobByRunnerExecutionID(ctx, e.JobExecutionID)
	if err != nil {
		if db.IsRecordNotFoundError(err) {
			log.ExtractLogger(ctx).Warningf("job not found for runner execution: %d", e.JobExecutionID)
			return nil
		}
		return errors.WithMessage(err, "failed to get job by runner execution id")
	}
	return h.refreshJob(ctx, job)
}

func (h *App) handleJobExecutionUpdate(ctx context.Context, e *api.JobExecution) error {
	job, err := h.db.GetJobByRunnerExecutionID(ctx, e.ID)
	if err != nil {
		if db.IsRecordNotFoundError(err) {
			log.ExtractLogger(ctx).Warningf("job not found for runner execution: %d", e.ID)
			return nil
		}
		return errors.WithMessage(err, "failed to get job by runner execution id")
	}
	// TODO: the status in refreshJob might be different from the status in the event.(db delay issue)
	return h.refreshJob(ctx, job)
}

func (h *App) RefreshJobExecution(ctx context.Context, id int64) error {
	job, err := h.db.GetJobByRunnerExecutionID(ctx, id)
	if err != nil {
		if db.IsRecordNotFoundError(err) {
			log.ExtractLogger(ctx).Warningf("job not found for runner execution: %d", id)
			return nil
		}
		return errors.WithMessage(err, "failed to get job by runner execution id")
	}
	return h.refreshJob(ctx, job)
}
