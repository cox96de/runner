package handler

import (
	"context"

	"github.com/cox96de/runner/api"

	"github.com/cox96de/runner/log"

	"github.com/cox96de/runner/db"
	"github.com/pkg/errors"
)

func (h *Handler) CreatePipeline(ctx context.Context, request *api.CreatePipelineRequest) (*api.CreatePipelineResponse, error) {
	logger := log.ExtractLogger(ctx)
	response, err := h.pipelineService.CreatePipeline(ctx, request.Pipeline)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create pipeline")
	}
	if err = h.dispatchService.Dispatch(ctx, response.CreatedJobs, response.CreatedJobExecutions); err != nil {
		logger.Warnf("failed to dispatch job: %+v", err)
	}
	p, err := packPipeline(response.CreatedPipeline, response.CreatedJobs, response.CreatedJobExecutions,
		response.CreatedSteps, response.CreatedStepExecutions)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to pack pipeline")
	}
	return &api.CreatePipelineResponse{
		Pipeline: p,
	}, nil
}

func packPipeline(p *db.Pipeline, jobs []*db.Job, jobExecutions []*db.JobExecution, steps []*db.Step,
	stepExecutions []*db.StepExecution,
) (*api.Pipeline, error) {
	pipeline := &api.Pipeline{
		ID:        p.ID,
		CreatedAt: api.ConvertTime(p.CreatedAt),
		UpdatedAt: api.ConvertTime(p.UpdatedAt),
	}
	stepsByJobID := make(map[int64][]*api.Step)
	stepExecutionByJobID := make(map[int64][]*db.StepExecution)
	for _, stepExecution := range stepExecutions {
		stepExecutionByJobID[stepExecution.JobExecutionID] = append(stepExecutionByJobID[stepExecution.JobExecutionID], stepExecution)
	}
	for _, step := range steps {
		s, err := db.PackStep(step, stepExecutionByJobID[step.ID])
		if err != nil {
			return nil, errors.WithMessagef(err, "failed to pack step %d", step.ID)
		}
		stepsByJobID[step.JobID] = append(stepsByJobID[step.JobID], s)
	}
	jobExecutionsByJobID := make(map[int64][]*db.JobExecution)
	for _, jobExecution := range jobExecutions {
		jobExecutionsByJobID[jobExecution.JobID] = append(jobExecutionsByJobID[jobExecution.JobID], jobExecution)
	}
	for _, job := range jobs {
		j, err := db.PackJob(job, jobExecutionsByJobID[job.ID], steps, stepExecutions)
		if err != nil {
			return nil, errors.WithMessagef(err, "failed to pack job %d", job.ID)
		}
		j.Steps = stepsByJobID[j.ID]
		pipeline.Jobs = append(pipeline.Jobs, j)
	}
	return pipeline, nil
}
