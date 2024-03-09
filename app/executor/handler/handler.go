package handler

import (
	"context"
	"sync"

	"github.com/cox96de/runner/app/executor/executorpb"
)

// Handler is the API set of the executor.
type Handler struct {
	executorpb.UnimplementedExecutorServer
	// commandLock to protect the commands map.
	commandLock sync.RWMutex
	// commands is a map of commandID to command.
	commands map[string]*command
}

// NewHandler creates a new Handler.
func NewHandler() *Handler {
	return &Handler{
		commands: make(map[string]*command),
	}
}

func (*Handler) Ping(context.Context, *executorpb.PingRequest) (*executorpb.PingResponse, error) {
	return &executorpb.PingResponse{}, nil
}
