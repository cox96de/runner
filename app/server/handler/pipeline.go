package handler

import (
	"context"

	"github.com/cox96de/runner/telemetry/trace"

	"github.com/cox96de/runner/api"

	"github.com/cox96de/runner/log"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/db"
)

func (h *Handler) CreatePipeline(ctx context.Context, request *api.CreatePipelineRequest) (*api.CreatePipelineResponse, error) {
	logger := log.ExtractLogger(ctx)
	ctx, span := trace.Start(ctx, "handler.create_pipeline")
	defer span.End()
	if err := api.ValidateDSL(request.Pipeline); err != nil {
		return nil, errors.WithMessage(err, "failed to validate pipeline DSL")
	}
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
	// job id -> step id -> step executions.
	stepExecutionMap := make(map[int64]map[int64][]*db.StepExecution)
	// JobExecutionID -> JobID
	jobExecutionID2JobID := make(map[int64]int64)
	for _, jobExecution := range jobExecutions {
		jobExecutionID2JobID[jobExecution.ID] = jobExecution.JobID
	}
	for _, stepExecution := range stepExecutions {
		jobID, ok := jobExecutionID2JobID[stepExecution.JobExecutionID]
		if !ok {
			return nil, errors.Errorf("step execution %d has no job", stepExecution.ID)
		}
		if stepExecutionMap[jobID] == nil {
			stepExecutionMap[jobID] = map[int64][]*db.StepExecution{}
		}
		stepExecutionMap[jobID][stepExecution.StepID] = append(stepExecutionMap[jobID][stepExecution.StepID], stepExecution)
	}
	for _, step := range steps {
		var stepExecutions []*db.StepExecution
		stepExecutionM := stepExecutionMap[step.JobID]
		if stepExecutionM != nil {
			stepExecutions = stepExecutionM[step.ID]
		}
		s, err := db.PackStep(step, stepExecutions)
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
		executions := jobExecutionsByJobID[job.ID]
		if len(executions) == 0 {
			return nil, errors.Errorf("job %d has no execution", job.ID)
		}
		latestJobExecution := executions[0]
		j, err := db.PackJob(job, latestJobExecution, nil, steps, stepExecutions)
		if err != nil {
			return nil, errors.WithMessagef(err, "failed to pack job %d", job.ID)
		}
		j.Steps = stepsByJobID[j.ID]
		pipeline.Jobs = append(pipeline.Jobs, j)
	}
	return pipeline, nil
}
