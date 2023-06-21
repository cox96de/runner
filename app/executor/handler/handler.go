package handler

import (
	"net/http"
	"sync"

	internalmodel "github.com/cox96de/runner/internal/model"
	"github.com/gin-gonic/gin"
)

// Handler is the API set of the executor.
type Handler struct {
	// commandLock to protect the commands map.
	commandLock sync.RWMutex
	// commands is a map of commandID to command.
	commands map[string]*command
}

// NewHandler creates a new Handler.
func NewHandler() *Handler {
	return &Handler{
		commands: map[string]*command{},
	}
}

// RegisterRoutes adds the routes to the gin engine.
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.UseRawPath = true
	r.Any("/ping", h.pingHandler)
	r.POST("/commands/:id", h.startCommandHandler)
	r.GET("/commands/:id/logs", h.getCommandLogHandler)
}

// pingHandler is a simple ping handler, uses to validate the executor is ready.
func (h *Handler) pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, &internalmodel.Message{
		Message: "pong",
	})
}
