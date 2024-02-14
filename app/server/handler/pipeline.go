package handler

import (
	"context"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/cox96de/runner/db"
	"github.com/cox96de/runner/entity"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/samber/lo"
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
	createJobOpts := make([]*db.CreateJobOption, 0, len(pipeline.Jobs))
	createJobExecutionOpts := make([]*db.CreateJobExecutionOption, 0, len(pipeline.Jobs))
	createStepOptMap := make(map[string][]*db.CreateStepOption)
	createStepExecutionOpts := make([]*db.CreateStepExecutionOption, 0)
	for _, job := range pipeline.Jobs {
		createJobOpts = append(createJobOpts, &db.CreateJobOption{
			Name:             job.Name,
			RunsOn:           job.RunsOn,
			WorkingDirectory: job.WorkingDirectory,
			EnvVar:           job.EnvVar,
			DependsOn:        job.DependsOn,
		})
		var stepOpts []*db.CreateStepOption
		for _, step := range job.Steps {
			stepOpts = append(stepOpts, &db.CreateStepOption{
				Name:             step.Name,
				User:             step.User,
				WorkingDirectory: step.WorkingDirectory,
				EnvVar:           step.EnvVar,
				DependsOn:        step.DependsOn,
				Commands:         step.Commands,
			})
		}
		createStepOptMap[job.Name] = stepOpts
	}
	var (
		createdPipeline       *db.Pipeline
		createdJobs           []*db.Job
		createdJobExeuctions  []*db.JobExecution
		createdSteps          []*db.Step
		createdStepExecutions []*db.StepExecution
		err                   error
	)
	err = h.db.Transaction(func(client *db.Client) error {
		createdPipeline, err = client.CreatePipeline(ctx)
		if err != nil {
			return errors.WithMessage(err, "failed to create pipeline")
		}
		lo.ForEach(createJobOpts, func(opt *db.CreateJobOption, i int) {
			opt.PipelineID = createdPipeline.ID
		})
		createdJobs, err = client.CreateJobs(ctx, createJobOpts)
		if err != nil {
			return errors.WithMessage(err, "failed to create jobs")
		}
		createStepOpts := make([]*db.CreateStepOption, 0)
		for _, createdJob := range createdJobs {
			stepOpts, ok := createStepOptMap[createdJob.Name]
			if !ok {
				return errors.Errorf("missing step options for job [%s]", createdJob.Name)
			}
			lo.ForEach(stepOpts, func(opt *db.CreateStepOption, i int) {
				opt.PipelineID = createdPipeline.ID
				opt.JobID = createdJob.ID
			})
			createStepOpts = append(createStepOpts, stepOpts...)

			createJobExecutionOpts = append(createJobExecutionOpts, &db.CreateJobExecutionOption{
				JobID:  createdJob.ID,
				Status: entity.JobStatusCreated,
			})
		}
		createdJobExeuctions, err = client.CreateJobExecutions(ctx, createJobExecutionOpts)
		if err != nil {
			return errors.WithMessage(err, "failed to create job executions")
		}
		jobExecutionByJobIDMap := lo.SliceToMap(createdJobExeuctions, func(item *db.JobExecution) (int64, *db.JobExecution) {
			return item.JobID, item
		})
		createdSteps, err = client.CreateSteps(ctx, createStepOpts)
		if err != nil {
			return errors.WithMessage(err, "failed to create steps")
		}
		for _, step := range createdSteps {
			createStepExecutionOpts = append(createStepExecutionOpts, &db.CreateStepExecutionOption{
				JobExecutionID: jobExecutionByJobIDMap[step.JobID].ID,
				StepID:         step.ID,
				Status:         entity.StepStatusCreated,
			})
		}
		createdStepExecutions, err = client.CreateStepExecutions(ctx, createStepExecutionOpts)
		if err != nil {
			return errors.WithMessage(err, "failed to create step executions")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	p, err := packPipeline(createdPipeline, createdJobs, createdJobExeuctions, createdSteps, createdStepExecutions)
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
