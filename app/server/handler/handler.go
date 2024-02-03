package handler

import (
	"net/http"

	"github.com/cox96de/runner/db"

	"github.com/gin-gonic/gin"
)

// Handler is the http handler for the server.
type Handler struct {
	db *db.Client
}

func NewHandler(db *db.Client) *Handler {
	return &Handler{db: db}
}

func (h *Handler) PingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
