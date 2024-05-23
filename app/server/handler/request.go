package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/cox96de/runner/api"

	"github.com/cox96de/runner/db"
	"github.com/cox96de/runner/lib"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (h *Handler) RequestJobHandler(c *gin.Context) {
	response, err := h.RequestJob(c, nil)
	if err != nil {
		log.Errorf("failed to request job: %v", err)
		c.JSON(http.StatusInternalServerError, &Message{Message: err})
		return
	}
	if response.Job != nil {
		JSON(c, http.StatusOK, response)
		return
	}
	JSON(c, http.StatusNoContent, response)
}

func (h *Handler) RequestJob(ctx context.Context, request *api.RequestJobRequest) (*api.RequestJobResponse, error) {
	// TODO: the limit should be configurable.
	jobExecutions, err := h.db.GetQueuedJobExecutions(ctx, 100)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get queued job executions")
	}
	for _, jobExecution := range jobExecutions {
		lock, err := h.locker.Lock(ctx, lib.BuildJobRequestLockKey(jobExecution.JobID), "request_job",
			time.Second*10)
		if err != nil {
			log.Warnf("failed to lock job execution: %v", err)
		}
		if !lock {
			continue
		}
		job, err := h.packJob(ctx, jobExecution)
		if err != nil {
			log.Warnf("failed to get job: %v", err)
			ok, err := h.locker.Unlock(ctx, lib.BuildJobRequestLockKey(jobExecution.JobID))
			if err != nil {
				log.Warnf("failed to unlock job execution: %v", err)
				continue
			}
			if !ok {
				log.Warnf("failed to unlock job execution")
			}
		}
		return &api.RequestJobResponse{
			Job: job,
		}, nil
	}
	return &api.RequestJobResponse{
		Job: nil,
	}, nil
}

func (h *Handler) packJob(ctx context.Context, jobExecution *db.JobExecution) (*api.Job, error) {
	job, err := h.db.GetJobByID(ctx, jobExecution.JobID)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get job")
	}
	steps, err := h.db.GetStepsByJobID(ctx, job.ID)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get steps")
	}
	stepExecutions, err := h.db.GetStepExecutionsByJobExecutionID(ctx, jobExecution.ID)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get step executions")
	}
	packJob, err := db.PackJob(job, jobExecution, nil, steps, stepExecutions)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to pack job")
	}
	return packJob, nil
}
