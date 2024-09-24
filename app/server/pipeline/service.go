package pipeline

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/db"
	"github.com/samber/lo"
)

type Service struct {
	dbClient *db.Client
}

func NewService(dbClient *db.Client) *Service {
	return &Service{dbClient: dbClient}
}

type CreatePipelineResponse struct {
	CreatedPipeline       *db.Pipeline
	CreatedJobs           []*db.Job
	CreatedJobExecutions  []*db.JobExecution
	CreatedSteps          []*db.Step
	CreatedStepExecutions []*db.StepExecution
}

// CreatePipeline creates a new pipeline and inserts into db.
func (s *Service) CreatePipeline(ctx context.Context, pipeline *api.PipelineDSL) (*CreatePipelineResponse, error) {
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
			Timeout:          job.Timeout,
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
		r   = &CreatePipelineResponse{}
		err error
	)
	err = s.dbClient.Transaction(func(client *db.Client) error {
		r.CreatedPipeline, err = client.CreatePipeline(ctx)
		if err != nil {
			return errors.WithMessage(err, "failed to create pipeline")
		}
		lo.ForEach(createJobOpts, func(opt *db.CreateJobOption, i int) {
			opt.PipelineID = r.CreatedPipeline.ID
		})
		r.CreatedJobs, err = client.CreateJobs(ctx, createJobOpts)
		if err != nil {
			return errors.WithMessage(err, "failed to create jobs")
		}
		createStepOpts := make([]*db.CreateStepOption, 0)
		for _, createdJob := range r.CreatedJobs {
			stepOpts, ok := createStepOptMap[createdJob.Name]
			if !ok {
				return errors.Errorf("missing step options for job [%s]", createdJob.Name)
			}
			lo.ForEach(stepOpts, func(opt *db.CreateStepOption, i int) {
				opt.PipelineID = r.CreatedPipeline.ID
				opt.JobID = createdJob.ID
			})
			createStepOpts = append(createStepOpts, stepOpts...)

			createJobExecutionOpts = append(createJobExecutionOpts, &db.CreateJobExecutionOption{
				JobID:  createdJob.ID,
				Status: api.StatusCreated,
			})
		}
		r.CreatedJobExecutions, err = client.CreateJobExecutions(ctx, createJobExecutionOpts)
		if err != nil {
			return errors.WithMessage(err, "failed to create job executions")
		}
		jobExecutionByJobIDMap := lo.SliceToMap(r.CreatedJobExecutions, func(item *db.JobExecution) (int64, *db.JobExecution) {
			return item.JobID, item
		})
		r.CreatedSteps, err = client.CreateSteps(ctx, createStepOpts)
		if err != nil {
			return errors.WithMessage(err, "failed to create steps")
		}
		for _, step := range r.CreatedSteps {
			createStepExecutionOpts = append(createStepExecutionOpts, &db.CreateStepExecutionOption{
				JobExecutionID: jobExecutionByJobIDMap[step.JobID].ID,
				StepID:         step.ID,
				Status:         api.StatusCreated,
			})
		}
		r.CreatedStepExecutions, err = client.CreateStepExecutions(ctx, createStepExecutionOpts)
		if err != nil {
			return errors.WithMessage(err, "failed to create step executions")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}
