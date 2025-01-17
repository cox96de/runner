package app

import (
	"context"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cox96de/runner/app/server"

	"github.com/cox96de/runner/githubapp/ghclient"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/app/server/handler"
	"github.com/cox96de/runner/githubapp/db"
	"github.com/cox96de/runner/githubapp/dsl"
	"github.com/cox96de/runner/log"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v64/github"
	"github.com/samber/lo"
)

const filePATH = ".xxci/ci.yaml"
const (
	CloneURLEnvKey = "CI_CLONE_URL"
	RefEnvKey      = "CI_REF"
)

type App struct {
	ghClient         *ghclient.Client
	runnerServer     *server.App
	baseURL          string
	db               *db.Client
	cloneStep        []string
	cloneStepWindows []string
	vms              map[string]*VMMeta
}

func (h *App) Send(ctx context.Context, event event.Event) protocol.Result {
	return h.handleCloudEvents(ctx, event)
}

func NewApp(ghClient *ghclient.Client, baseURL string, dbCli *db.Client, cloneStep []string, cloneStepWindows []string,
	vms []*VMMeta,
) *App {
	return &App{
		ghClient:         ghClient,
		baseURL:          baseURL,
		db:               dbCli,
		cloneStep:        cloneStep,
		cloneStepWindows: cloneStepWindows,
		vms: lo.SliceToMap(vms, func(item *VMMeta) (string, *VMMeta) {
			return item.Image, item
		}),
	}
}

func (h *App) SetRunnerServer(runnerServer *server.App) {
	h.runnerServer = runnerServer
}

func (h *App) Handles() []string {
	return []string{"check_run", "check_suite"}
}

func (h *App) Handle(ctx context.Context, eventType, deliveryID string, payload []byte) error {
	logger := log.ExtractLogger(ctx).WithField("delivery.id", deliveryID)
	switch eventType {
	case "check_run", "check_suite", "installation":
		hook, err := github.ParseWebHook(eventType, payload)
		logger.Debugf("handle event:%s, %s", eventType, string(payload))
		if err != nil {
			return errors.WithMessage(err, "failed to parse webhook")
		}
		switch event := hook.(type) {
		case *github.CheckRunEvent:
			if event.GetAction() != "rerequested" {
				return nil
			}
			return h.handleRerequestCheckRun(ctx, event)
		case *github.CheckSuiteEvent:
			if err := h.handleCheckSuite(ctx, event); err != nil {
				log.ExtractLogger(ctx).Errorf("failed to handle check suite event: %v", err)
				return errors.WithMessage(err, "failed to handle check suite event")
			}
		case *github.InstallationEvent:
			logger.Debugf("handle event: %+v", event)
		default:
			return errors.Errorf("unsupported event type: %T", event)
		}
	default:
		return nil
	}
	return nil
}

func (h *App) handleCheckSuite(ctx context.Context, event *github.CheckSuiteEvent) error {
	logger := log.ExtractLogger(ctx)
	if event.Action == nil || *event.Action != "requested" {
		logger.Infof("action is not requested: %s", *event.Action)
		return nil
	}
	client, err := h.ghClient.AppInstallClient(event.Installation.GetID())
	if err != nil {
		return errors.WithMessage(err, "failed to create github client")
	}
	headSHA := event.CheckSuite.GetHeadSHA()
	content, _, resp, err := client.Repositories.GetContents(ctx, event.Repo.Owner.GetLogin(), event.Repo.GetName(), filePATH, &github.RepositoryContentGetOptions{
		Ref: headSHA,
	})
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			logger.Infof("no ci file is found")
			return nil
		}
		return errors.WithMessage(err, "failed to get repository content")
	}
	cc, err := content.GetContent()
	if err != nil {
		return errors.WithMessage(err, "failed to get content")
	}
	pipelineDSL, err := dsl.ParseFromContent([]byte(cc))
	if err != nil {
		return errors.WithMessage(err, "failed to parse pipeline")
	}
	// Append clone step.
	for jobID, job := range pipelineDSL.Jobs {
		cloneStep := h.cloneStep
		meta, ok := h.getImageMeta(job.RunsOn.Image)
		if ok && meta.OS == api.WindowsOS {
			cloneStep = h.cloneStepWindows
		}
		job.Steps = append([]*dsl.Step{
			{
				Name: "clone",
				Run:  cloneStep,
				Env: map[string]string{
					CloneURLEnvKey: event.Repo.GetCloneURL(),
					RefEnvKey:      event.CheckSuite.GetAfterSHA(),
				},
			},
		}, job.Steps...)
		pipelineDSL.Jobs[jobID] = job
	}
	checkRun, response, err := client.Checks.CreateCheckRun(ctx, event.Repo.Owner.GetLogin(), event.Repo.GetName(), github.CreateCheckRunOptions{
		Name:       "[Welcome]",
		HeadSHA:    headSHA,
		Status:     lo.ToPtr("completed"),
		Conclusion: lo.ToPtr("success"),
		StartedAt: &github.Timestamp{
			Time: time.Now(),
		},
		CompletedAt: &github.Timestamp{
			Time: time.Now(),
		},
		Output:  nil,
		Actions: nil,
	})
	if err != nil {
		return errors.WithMessagef(err, "failed to create check run: %v", response)
	}
	logger.Infof("created check run: %d", checkRun.GetID())
	pipelinePO, err := h.db.CreatePipeline(ctx, &db.CreatePipelineOption{
		AppInstallID: event.Installation.GetID(),
		RepoOwner:    event.Repo.Owner.GetLogin(),
		RepoName:     event.Repo.GetName(),
		HeadSHA:      headSHA,
	})
	if err != nil {
		return errors.WithMessage(err, "failed to insert pipeline")
	}
	var runnerPipeline *api.Pipeline
	if runnerPipeline, err = h.createPipeline(ctx, event.Repo.GetName(), pipelineDSL); err != nil {
		return errors.WithMessage(err, "failed to create pipeline")
	}
	runnerJobMap := lo.SliceToMap(runnerPipeline.Jobs, func(item *api.Job) (string, *api.Job) {
		return item.Name, item
	})
	createJobOptions := make([]*db.CreateJobOption, 0, len(runnerJobMap))
	for jobID, job := range pipelineDSL.Jobs {
		runnerJob := runnerJobMap[jobID]
		createCheckRunOptions, err := GenerateCreateCheckRunOptions(h.baseURL, headSHA, &RenderCheckRunOptions{
			RunnerJob: runnerJob.Execution,
			RenderJob: &RenderJob{
				UID:  jobID,
				Name: job.Name,
				Steps: lo.Map(job.Steps, func(step *dsl.Step, _ int) *RenderStep {
					return &RenderStep{Name: step.Name}
				}),
			},
		})
		if err != nil {
			return errors.WithMessage(err, "failed to render check run")
		}
		checkRun, _, err = client.Checks.CreateCheckRun(ctx, event.Repo.Owner.GetLogin(), event.Repo.GetName(), createCheckRunOptions)
		if err != nil {
			return errors.WithMessagef(err, "failed to create check run for job: %s", jobID)
		}
		createJobOptions = append(createJobOptions, &db.CreateJobOption{
			PipelineID:           pipelinePO.ID,
			Name:                 job.Name,
			UID:                  jobID,
			CheckRunID:           checkRun.GetID(),
			RunnerJobExecutionID: runnerJob.Execution.JobID,
			Steps: lo.Map(job.Steps, func(step *dsl.Step, _ int) *db.Step {
				return &db.Step{Name: step.Name}
			}),
		})
	}
	_, err = h.db.CreateJobs(ctx, createJobOptions)
	if err != nil {
		return errors.WithMessage(err, "failed to insert jobs")
	}
	return nil
}

func (h *App) handleRerequestCheckRun(ctx context.Context, event *github.CheckRunEvent) error {
	logger := log.ExtractLogger(ctx)
	job, err := h.db.GetJobByCheckRunID(ctx, event.GetCheckRun().GetID())
	if err != nil {
		return errors.WithMessage(err, "failed to get job by check run id")
	}
	getJobExecutionResponse, err := h.runnerServer.GetJobExecution(ctx, &api.GetJobExecutionRequest{
		JobExecutionID: job.RunnerJobExecutionID,
	})
	if err != nil {
		return errors.WithMessage(err, "failed to get job execution")
	}
	rerunJobResponse, err := h.runnerServer.RerunJob(ctx, &api.RerunJobRequest{
		JobID: getJobExecutionResponse.JobExecution.JobID,
	})
	if err != nil {
		return errors.WithMessage(err, "failed to rerun job")
	}
	client, err := h.ghClient.AppInstallClient(event.Installation.GetID())
	if err != nil {
		return errors.WithMessage(err, "failed to get app install client")
	}
	oldCheckRun, _, err := client.Checks.GetCheckRun(ctx, event.GetRepo().GetOwner().GetLogin(), event.GetRepo().GetName(), job.CheckRunID)
	if err != nil {
		return errors.WithMessage(err, "failed to get old check run")
	}
	checkRun, _, err := client.Checks.CreateCheckRun(ctx, event.GetRepo().GetOwner().GetLogin(), event.GetRepo().GetName(), github.CreateCheckRunOptions{
		Name:       lo.FromPtr(oldCheckRun.Name),
		HeadSHA:    lo.FromPtr(oldCheckRun.HeadSHA),
		DetailsURL: oldCheckRun.DetailsURL,
		ExternalID: oldCheckRun.ExternalID,
		Status:     lo.ToPtr("queued"),
	})
	if err != nil {
		return errors.WithMessage(err, "failed to create new check run")
	}
	logger.Infof("created new check run with id: %d", checkRun.GetID())
	if err := h.db.UpdateJob(ctx, job.ID, &db.UpdateJobOption{
		RunnerJobExecutionID: &rerunJobResponse.JobExecution.ID,
		CheckRunID:           checkRun.ID,
	}); err != nil {
		return errors.WithMessage(err, "failed to update job")
	}
	job.RunnerJobExecutionID = rerunJobResponse.JobExecution.ID
	job.CheckRunID = checkRun.GetID()
	return nil
}

func (h *App) createPipeline(ctx context.Context, repoName string, p *dsl.Pipeline) (*api.Pipeline, error) {
	runnerPipeline := &api.PipelineDSL{}
	workdir := filepath.Join("/home/runner/work/", repoName)
	for jobID, job := range p.Jobs {
		runnerSteps := make([]*api.StepDSL, 0, len(job.Steps))
		for idx, step := range job.Steps {
			stepDSL := &api.StepDSL{
				Name:             strconv.Itoa(idx),
				Commands:         step.Run,
				WorkingDirectory: workdir,
				EnvVar:           step.Env,
			}
			runnerSteps = append(runnerSteps, stepDSL)
		}
		on := job.RunsOn
		var runsOn *api.RunsOn
		switch {
		case on == nil:
			return nil, errors.Errorf("runs_on is required for job: %s", jobID)
		case on.ContainerImage != "":
			mainContainer := "runner"
			runsOn = &api.RunsOn{
				// HARDCODE: the label is hardcoded.
				Label: "kube",
				Docker: &api.Docker{
					Containers: []*api.Container{
						{
							Name:  mainContainer,
							Image: on.ContainerImage,
						},
					},
					DefaultContainer: mainContainer,
				},
			}
		case on.Image != "":
			os := "linux"
			meta, ok := h.getImageMeta(on.Image)
			if ok {
				os = meta.OS
			}
			runsOn = &api.RunsOn{
				Label: "vm",
				VM: &api.VM{
					Image:  on.Image,
					OS:     os,
					CPU:    4,
					Memory: 8196,
				},
			}
		default:
			return nil, errors.Errorf("runs_on should be not empty: %s", jobID)
		}
		runnerPipeline.Jobs = append(runnerPipeline.Jobs, &api.JobDSL{
			Name:   jobID,
			RunsOn: runsOn,
			Steps:  runnerSteps,
		})
	}
	pipeline, err := h.runnerServer.CreatePipeline(ctx, &api.CreatePipelineRequest{
		Pipeline: runnerPipeline,
	})
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create pipeline")
	}
	log.ExtractLogger(ctx).Infof("created pipeline: %d", pipeline.Pipeline.ID)
	return pipeline.Pipeline, nil
}

type GetLogRequest struct {
	JobExecutionID int64  `path:"job_execution_id"`
	LogName        string `path:"log_name"`
	Offset         int64  `query:"offset"`
}

type GetLogResponse struct {
	Logs []*api.LogLine `json:"logs"`
}

func (h *App) GetLogHandler(ctx *gin.Context) {
	req := &GetLogRequest{}
	if err := handler.Bind(ctx, req); err != nil {
		handler.JSON(ctx, http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	getLogLinesResponse, err := h.runnerServer.GetLogLines(ctx, &api.GetLogLinesRequest{
		JobExecutionID: req.JobExecutionID,
		Name:           req.LogName,
		Offset:         req.Offset,
		Limit:          lo.ToPtr(int64(100)),
	})
	if err != nil {
		handler.JSON(ctx, http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	handler.JSON(ctx, http.StatusOK, &GetLogResponse{Logs: getLogLinesResponse.Lines})
}
