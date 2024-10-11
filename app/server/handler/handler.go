package handler

import (
	"context"

	"github.com/cox96de/runner/app/server/eventhook"

	"github.com/cox96de/runner/app/server/logstorage"

	"github.com/cox96de/runner/api"

	"github.com/cox96de/runner/app/server/dispatch"
	"github.com/cox96de/runner/app/server/pipeline"
	"github.com/cox96de/runner/lib"

	"github.com/cox96de/runner/db"
)

var _ api.ServerServer = (*Handler)(nil)

// Handler is the http handler for the server.
type Handler struct {
	api.UnimplementedServerServer
	db              *db.Client
	pipelineService *pipeline.Service
	dispatchService *dispatch.Service
	logService      *logstorage.Service
	locker          lib.Locker
	eventhook       *eventhook.Service
}

// nolint: unused
func (h *Handler) mustEmbedUnimplementedServerServer() {
}

func NewHandler(db *db.Client, pipelineService *pipeline.Service, dispatchService *dispatch.Service,
	locker lib.Locker, logService *logstorage.Service, eventhook *eventhook.Service,
) *Handler {
	return &Handler{
		db: db, pipelineService: pipelineService, dispatchService: dispatchService, locker: locker,
		logService: logService, eventhook: eventhook,
	}
}

func (h *Handler) Ping(ctx context.Context, _ *api.ServerPingRequest) (*api.ServerPingResponse, error) {
	return &api.ServerPingResponse{}, nil
}
