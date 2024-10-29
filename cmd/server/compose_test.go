package main

import (
	"context"
	"net/http/httptest"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/gin-gonic/gin"
	"gotest.tools/v3/assert"
)

func TestComposeCloudEventsClient(t *testing.T) {
	t.Run("http", func(t *testing.T) {
		engine := gin.New()
		proto, err := cloudevents.NewHTTP()
		assert.NilError(t, err)
		reciever := func(event cloudevents.Event) error {
			assert.Equal(t, string(event.Data()), "{\"hello\":\"world\"}")
			return nil
		}
		ceh, err := cloudevents.NewHTTPReceiveHandler(context.Background(), proto, reciever)
		assert.NilError(t, err)
		engine.POST("/cloudevents", func(c *gin.Context) {
			ceh.ServeHTTP(c.Writer, c.Request)
		})
		server := httptest.NewServer(engine)
		client, err := ComposeCloudEventsClient(&Event{HTTPEndPoint: server.URL + "/cloudevents"})
		assert.NilError(t, err)
		event := cloudevents.NewEvent()
		event.SetSource("runner")
		event.SetType("test")
		err = event.SetData(cloudevents.ApplicationJSON, map[string]string{"hello": "world"})
		assert.NilError(t, err)
		err = client.Send(context.Background(), event)
		assert.Assert(t, !cloudevents.IsUndelivered(err))
	})
	t.Run("nop", func(t *testing.T) {
		client, err := ComposeCloudEventsClient(nil)
		assert.NilError(t, err)
		event := cloudevents.NewEvent()
		err = client.Send(context.Background(), event)
		assert.NilError(t, err)
	})
}
