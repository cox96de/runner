package handler

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) RegisterRouter(g *gin.RouterGroup) {
	g.POST("/ping", h.PingHandler)
	g.POST("/pipelines", h.CreatePipelineHandler)
	g.POST("/jobs/request", h.RequestJobHandler)
	g.POST("/jobs/:job_id/executions/:id", getGinHandler(h.UpdateJobExecution))
}
