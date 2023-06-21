package handler

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/cox96de/runner/internal/executor"
	"github.com/gin-gonic/gin"
	"gotest.tools/v3/assert"
)

func TestHandler_pingHandler(t *testing.T) {
	engine := gin.New()
	handler := NewHandler()
	handler.RegisterRoutes(engine)
	testServer := httptest.NewServer(engine)
	defer testServer.Close()
	client := executor.NewClient(testServer.URL)
	err := client.Ping(context.Background())
	assert.NilError(t, err)
}
