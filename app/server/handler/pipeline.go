package handler

import (
	"context"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/cox96de/runner/db"
	"github.com/cox96de/runner/entity"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type CreatePipelineRequest struct {
	// Pipeline is pipeline DSL.
	Pipeline *entity.Pipeline `json:"pipeline"`
}

type CreatePipelineResponse struct {
	Pipeline *entity.Pipeline `json:"pipeline"`
}

func (h *Handler) CreatePipelineHandler(c *gin.Context) {
	var req CreatePipelineRequest
	if err := Bind(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, &Message{Message: err})
		return
	}
	pipeline, err := h.createPipeline(c, req.Pipeline)
	if err != nil {
		log.Errorf("failed to create pipeline: %v", err)
		c.JSON(http.StatusInternalServerError, &Message{Message: err})
		return
	}
	c.JSON(http.StatusOK, &CreatePipelineResponse{Pipeline: pipeline})
}

func (h *Handler) createPipeline(ctx context.Context, pipeline *entity.Pipeline) (*entity.Pipeline, error) {
	response, err := h.pipelineService.CreatePipeline(ctx, pipeline)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create pipeline")
	}
	if err = h.dispatchService.Dispatch(ctx, response.CreatedJobs, response.CreatedJobExecutions); err != nil {
		log.Warnf("failed to dispatch job: %+v", err)
	}
	p, err := packPipeline(response.CreatedPipeline, response.CreatedJobs, response.CreatedJobExecutions,
		response.CreatedSteps, response.CreatedStepExecutions)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to pack pipeline")
	}
	return p, nil
}

func packPipeline(p *db.Pipeline, jobs []*db.Job, jobExecutions []*db.JobExecution, steps []*db.Step,
	stepExecutions []*db.StepExecution,
) (*entity.Pipeline, error) {
	pipeline := &entity.Pipeline{
		ID:        p.ID,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	stepsByJobID := make(map[int64][]*entity.Step)
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
		j, err := db.PackJob(job, jobExecutionsByJobID[job.ID])
		if err != nil {
			return nil, errors.WithMessagef(err, "failed to pack job %d", job.ID)
		}
		j.Steps = stepsByJobID[j.ID]
		pipeline.Jobs = append(pipeline.Jobs, j)
	}
	return pipeline, nil
}
