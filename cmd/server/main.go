package main

import (
	"github.com/cox96de/runner/app/server/handler"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func main() {
	h := handler.NewHandler()
	engine := gin.New()
	group := engine.Group("/api/v1")
	h.RegisterRouter(group)
	// TODO: load port from config.
	err := engine.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
