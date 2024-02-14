package handler

import (
	"net/http"

	"github.com/cox96de/runner/app/server/dispatch"
	"github.com/cox96de/runner/app/server/pipeline"

	"github.com/cox96de/runner/db"

	"github.com/gin-gonic/gin"
)

// Handler is the http handler for the server.
type Handler struct {
	db              *db.Client
	pipelineService *pipeline.Service
	dispatchService *dispatch.Service
}

func NewHandler(db *db.Client, pipelineService *pipeline.Service, dispatchService *dispatch.Service) *Handler {
	return &Handler{db: db, pipelineService: pipelineService, dispatchService: dispatchService}
}

func (h *Handler) PingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
