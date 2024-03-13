package handler

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) RegisterRouter(g *gin.RouterGroup) {
	g.POST("/ping", h.PingHandler)
	g.POST("/pipelines", getGinHandler(h.CreatePipeline))
	g.POST("/jobs/request", h.RequestJobHandler)
	g.POST("/jobs/:job_id/executions/:job_execution_id", getGinHandler(h.UpdateJobExecution))
	g.POST("/jobs/:job_id/executions/:job_execution_id/logs", getGinHandler(h.UploadLogLines))
	g.GET("/jobs/:job_id/executions/:job_execution_id/logs/:name", getGinHandler(h.GetLogLines))
	g.POST("/jobs/:job_id/executions/:job_execution_id/steps/:step_execution_id", getGinHandler(h.UpdateStepExecution))
}
