package executor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	internalmodel "github.com/cox96de/runner/internal/model"
	"github.com/gin-gonic/gin"
	"gotest.tools/v3/assert"
)

func TestClient_Ping(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		engine := gin.New()
		engine.Any(pingEndpoint, func(c *gin.Context) {
			c.JSON(http.StatusOK, &internalmodel.Message{Message: "pong"})
		})
		server := httptest.NewServer(engine)
		defer server.Close()
		client := NewClient(server.URL)
		err := client.Ping(context.Background())
		assert.NilError(t, err)
	})
	t.Run("bad", func(t *testing.T) {
		engine := gin.New()
		message := "not found"
		engine.Any(pingEndpoint, func(c *gin.Context) {
			c.JSON(http.StatusNotFound, &internalmodel.Message{Message: message})
		})
		server := httptest.NewServer(engine)
		defer server.Close()
		client := NewClient(server.URL)
		err := client.Ping(context.Background())
		assert.ErrorContains(t, err, message)
	})
	t.Run("connection_refuse", func(t *testing.T) {
		client := NewClient("http://localhost:1234")
		err := client.Ping(context.Background())
		if runtime.GOOS == "windows" {
			assert.ErrorContains(t, err, "target machine actively refused it")
		} else {
			assert.ErrorContains(t, err, "connection refused")
		}
	})
}
