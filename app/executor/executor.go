package executor

import (
	"net"

	"github.com/cox96de/runner/app/executor/executorpb"
	"google.golang.org/grpc"

	"github.com/cox96de/runner/app/executor/handler"
)

type App struct {
	handler *handler.Handler
	server  *grpc.Server
}

func NewApp() *App {
	app := &App{
		handler: handler.NewHandler(),
		server:  grpc.NewServer(),
	}
	executorpb.RegisterExecutorServer(app.server, app.handler)
	return app
}

// Run starts the server.
func (app *App) Run(listener net.Listener) error {
	return app.server.Serve(listener)
}
