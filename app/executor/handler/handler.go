package handler

import (
	internalmodel "github.com/cox96de/runner/internal/model"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Handler is the API set of the executor.
type Handler struct {
}

// NewHandler creates a new Handler.
func NewHandler() *Handler {
	return &Handler{}
}

// RegisterRoutes adds the routes to the gin engine.
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.UseRawPath = true
	r.Any("/ping", h.pingHandler)
}

// pingHandler is a simple ping handler, uses to validate the executor is ready.
func (h *Handler) pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, &internalmodel.Message{
		Message: "pong",
	})
}
