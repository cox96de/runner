package eventhook

import (
	"context"

	"github.com/cox96de/runner/log"

	cloudeventshttp "github.com/cloudevents/sdk-go/v2/protocol/http"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/api"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cox96de/runner/db"
)

const source = "runner-server"

type nopSender struct{}

func (n *nopSender) Send(_ context.Context, _ event.Event) protocol.Result {
	return nil
}

func NewNopSender() Sender {
	return &nopSender{}
}

type Sender interface {
	Send(ctx context.Context, event event.Event) protocol.Result
}

type Service struct {
	client Sender
}

func NewService(client Sender) *Service {
	return &Service{client: client}
}

// SendStepExecutionEvent sends a step execution event to the event hook.
func (s *Service) SendStepExecutionEvent(ctx context.Context, step *db.StepExecution) error {
	ev := cloudevents.NewEvent()
	ev.SetType("step_execution")
	if err := ev.SetData(cloudevents.ApplicationJSON, api.Event{
		StepExecution: db.PackStepExecution(step),
	}); err != nil {
		return errors.WithMessage(err, "failed to set data")
	}
	ctx = log.WithLogger(ctx, log.ExtractLogger(ctx).WithField("step_id", step.ID))
	go s.doSend(ctx, ev)
	return nil
}

// SendJobExecutionEvent sends a job execution event to the event hook.
func (s *Service) SendJobExecutionEvent(ctx context.Context, job *db.JobExecution) error {
	ev := cloudevents.NewEvent()
	ev.SetType("step_execution")
	jobExecution, err := db.PackJobExecution(job, nil)
	if err != nil {
		return errors.WithMessagef(err, "failed to pack job execution")
	}
	if err := ev.SetData(cloudevents.ApplicationJSON, api.Event{
		JobExecution: jobExecution,
	}); err != nil {
		return errors.WithMessage(err, "failed to set data")
	}
	ctx = log.WithLogger(ctx, log.ExtractLogger(ctx).WithField("step_id", jobExecution.ID))
	go s.doSend(ctx, ev)
	return nil
}

func (s *Service) doSend(ctx context.Context, ev event.Event) {
	logger := log.ExtractLogger(ctx).WithFields(log.Fields{
		"event_id": ev.ID(),
		"type":     ev.Type(),
	})
	ev.SetSource(source)
	result := s.client.Send(log.WithLogger(context.Background(), logger), ev)
	var httpResult *cloudeventshttp.Result
	switch {
	case cloudevents.IsACK(result):
		logger.Infof("event is deliveried")
		return
	case cloudevents.ResultAs(result, &httpResult):
		logger.Errorf("http result with code: %d, %+v", httpResult.StatusCode, httpResult)
		return
	case cloudevents.IsNACK(result):
		logger.Infof("event is not ack")
		return

	}
	// Omit error because: don't abort process by event hook.
	// Persistent event and resent later.
	logger.Errorf("failed to send event: %+v", result)
}
