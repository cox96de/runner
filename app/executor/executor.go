package executor

import (
	"fmt"
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
		server: r,
	}

	h := handler.NewHandler()

	h.RegisterRoutes(r)

	return app
}

// Run starts the server.
func (app *App) Run(port int) error {
	return app.server.Run(fmt.Sprintf(":%d", port))
}
