package handler

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/cox96de/runner/internal/executor"
	"github.com/gin-gonic/gin"
	"gotest.tools/v3/assert"
)

func setupHandler(t *testing.T) (*httptest.Server, *Handler) {
	engine := gin.New()
	handler := NewHandler()
	handler.RegisterRoutes(engine)
	testServer := httptest.NewServer(engine)
	t.Cleanup(func() {
		testServer.Close()
	})
	return testServer, handler
}

func TestHandler_pingHandler(t *testing.T) {
	testServer, _ := setupHandler(t)
	client := executor.NewClient(testServer.URL)
	err := client.Ping(context.Background())
	assert.NilError(t, err)
}
