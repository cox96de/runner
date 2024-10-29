package eventhook

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cloudeventshttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/db"
	"github.com/gin-gonic/gin"
	"gotest.tools/v3/assert"
)

func TestService_SendStepExecutionEvent(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		stepExecution := &db.StepExecution{
			ID:             1,
			JobExecutionID: 2,
			StepID:         3,
			Status:         api.StatusRunning,
			ExitCode:       0,
		}
		engine := gin.New()
		server := httptest.NewServer(engine)
		proto, err := cloudevents.NewHTTP()
		assert.NilError(t, err)
		e := &api.Event{}
		handleCh := make(chan struct{})
		handler, err := cloudevents.NewHTTPReceiveHandler(context.Background(), proto, func(ctx context.Context, event cloudevents.Event) error {
			assert.Assert(t, len(event.ID()) > 0)
			err := event.DataAs(e)
			assert.NilError(t, err)
			close(handleCh)
			return nil
		})
		assert.NilError(t, err)
		engine.POST("/event", gin.WrapH(handler))
		proto, err = cloudevents.NewHTTP(cloudevents.WithTarget(server.URL + "/event"))
		assert.NilError(t, err)
		http, err := cloudevents.NewClient(proto, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
		assert.NilError(t, err)
		service := NewService(http)
		err = service.SendStepExecutionEvent(context.Background(), stepExecution)
		assert.NilError(t, err)
		<-handleCh
		assert.DeepEqual(t, e.StepExecution.StepID, stepExecution.StepID)
		assert.DeepEqual(t, e.StepExecution.ID, stepExecution.ID)
	})
	t.Run("bad", func(t *testing.T) {
		stepExecution := &db.StepExecution{
			ID:             1,
			JobExecutionID: 2,
			StepID:         3,
			Status:         api.StatusRunning,
			ExitCode:       0,
		}
		engine := gin.New()
		server := httptest.NewServer(engine)
		proto, err := cloudevents.NewHTTP()
		assert.NilError(t, err)
		handleCh := make(chan struct{})
		handler, err := cloudevents.NewHTTPReceiveHandler(context.Background(), proto, func(ctx context.Context, event cloudevents.Event) error {
			defer close(handleCh)
			return cloudeventshttp.NewResult(http.StatusConflict, "a format: %", event)
		})
		assert.NilError(t, err)
		engine.POST("/event_error", gin.WrapH(handler))
		proto, err = cloudevents.NewHTTP(cloudevents.WithTarget(server.URL + "/event_error"))
		assert.NilError(t, err)
		http, err := cloudevents.NewClient(proto, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
		assert.NilError(t, err)
		service := NewService(http)
		err = service.SendStepExecutionEvent(context.Background(), stepExecution)
		assert.NilError(t, err)
		<-handleCh
		time.Sleep(time.Millisecond * 10)
	})
}

func TestService_SendJobExecutionEvent(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		jobExecution := &db.JobExecution{
			ID:     1,
			Status: api.StatusRunning,
		}
		engine := gin.New()
		server := httptest.NewServer(engine)
		proto, err := cloudevents.NewHTTP()
		assert.NilError(t, err)
		e := &api.Event{}
		handleCh := make(chan struct{})
		handler, err := cloudevents.NewHTTPReceiveHandler(context.Background(), proto, func(ctx context.Context, event cloudevents.Event) error {
			assert.Assert(t, len(event.ID()) > 0)
			err := event.DataAs(e)
			assert.NilError(t, err)
			close(handleCh)
			return nil
		})
		assert.NilError(t, err)
		engine.POST("/event", gin.WrapH(handler))
		proto, err = cloudevents.NewHTTP(cloudevents.WithTarget(server.URL + "/event"))
		assert.NilError(t, err)
		http, err := cloudevents.NewClient(proto, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
		assert.NilError(t, err)
		service := NewService(http)
		err = service.SendJobExecutionEvent(context.Background(), jobExecution)
		assert.NilError(t, err)
		<-handleCh
		assert.DeepEqual(t, e.JobExecution.ID, jobExecution.ID)
	})
}
