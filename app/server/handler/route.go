package handler

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) RegisterRouter(g *gin.RouterGroup) {
	g.Any("/ping", getGinHandler(h.Ping))
	g.POST("/pipelines", getGinHandler(h.CreatePipeline))
	g.POST("/jobs/request", h.RequestJobHandler)
	g.GET("/jobs/:job_id/executions/", getGinHandler(h.ListJobExecutions))
	g.POST("/job_executions/:job_execution_id", getGinHandler(h.UpdateJobExecution))
	g.POST("/job_executions/:job_execution_id/heartbeat", getGinHandler(h.Heartbeat))
	g.GET("/job_executions/:job_execution_id", getGinHandler(h.GetJobExecution))
	g.POST("/job_executions/:job_execution_id/logs", getGinHandler(h.UploadLogLines))
	g.GET("/job_executions/:job_execution_id/logs/:name", getGinHandler(h.GetLogLines))
	g.GET("/step_executions/:step_execution_id", getGinHandler(h.GetStepExecution))
	g.POST("/step_executions/:step_execution_id", getGinHandler(h.UpdateStepExecution))
}
