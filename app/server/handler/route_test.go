package handler

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandler_RegisterRouter(t *testing.T) {
	h := &Handler{}
	g := gin.New()
	h.RegisterRouter(g.Group("/"))
}
