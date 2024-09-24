package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/cox96de/runner/api"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/db"
	"github.com/cox96de/runner/lib"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (h *Handler) RequestJobHandler(c *gin.Context) {
	request := &api.RequestJobRequest{}
	if err := Bind(c, request); err != nil {
		JSON(c, http.StatusBadRequest, &Message{Message: err})
		return
	}
	response, err := h.RequestJob(c, request)
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
	jobQueues, err := h.db.GetQueuedJobExecutionIDs(ctx, request.Label, 100)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get queued job executions")
	}
	for _, jobQueue := range jobQueues {
		job, ok := h.tryLockJobExecution(ctx, jobQueue.JobExecutionID)
		if !ok {
			continue
		}
		log.WithContext(ctx).Infof("dispatch job: %v, job execution: %d", job.ID, jobQueue.JobExecutionID)
		return &api.RequestJobResponse{
			Job: job,
		}, nil
	}
	return &api.RequestJobResponse{
		Job: nil,
	}, nil
}

func (h *Handler) tryLockJobExecution(ctx context.Context, jobExecutionID int64) (job *api.Job, ok bool) {
	lockKey := lib.BuildJobRequestLockKey(jobExecutionID)
	lock, err := h.locker.Lock(ctx, lockKey, "request_job",
		time.Second*10)
	if err != nil {
		log.Warnf("failed to lock job execution: %v", err)
		return nil, false
	}
	if !lock {
		return nil, false
	}
	defer func() {
		// If success to fetch the job, don't need to unlock it.
		if !ok {
			_, err := h.locker.Unlock(ctx, lockKey)
			if err != nil {
				log.Warnf("failed to unlock job execution: %v", err)
			}
		}
	}()
	jobExecution, err := h.db.GetJobExecution(ctx, jobExecutionID)
	if err != nil {
		log.Errorf("failed to get job execution: %v", err)
		return nil, false
	}
	job, err = h.packJob(ctx, jobExecution)
	if err != nil {
		log.Errorf("failed to pack job: %v", err)
		return nil, false
	}
	return job, true
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
