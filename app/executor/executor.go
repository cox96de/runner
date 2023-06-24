package executor

import (
	"net"

	"github.com/cox96de/runner/app/executor/handler"
	"github.com/gin-gonic/gin"
)

type App struct {
	server  *gin.Engine
	handler *handler.Handler
}

func NewApp() *App {
	r := gin.Default()
	app := &App{
		server:  r,
		handler: handler.NewHandler(),
	}
	app.handler.RegisterRoutes(r)
	return app
}

// Run starts the server.
func (app *App) Run(listener net.Listener) error {
	return app.server.RunListener(listener)
}
