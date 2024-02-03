package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler is the http handler for the server.
type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) PingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
