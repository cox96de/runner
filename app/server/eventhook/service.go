package eventhook

import (
	"context"

	"github.com/cox96de/runner/db"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

// SendStepExecutionEvent sends a step execution event to the event hook.
func (s *Service) SendStepExecutionEvent(ctx context.Context, step *db.StepExecution) error {
	return nil
}

// SendJobExecutionEvent sends a job execution event to the event hook.
func (s *Service) SendJobExecutionEvent(ctx context.Context, job *db.JobExecution) error {
	return nil
}
