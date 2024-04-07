package handler

import (
	"context"
	"runtime"
	"testing"

	"github.com/cox96de/runner/app/executor/executorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gotest.tools/v3/assert"
)

func TestHandler_GetRuntimeInfo(t *testing.T) {
	_, addr := setupHandler(t)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NilError(t, err)
	client := executorpb.NewExecutorClient(conn)
	getRuntimeInfoResp, err := client.GetRuntimeInfo(context.Background(), &executorpb.GetRuntimeInfoRequest{})
	assert.NilError(t, err)
	assert.Equal(t, getRuntimeInfoResp.OS, runtime.GOOS)
	assert.Equal(t, getRuntimeInfoResp.Arch, runtime.GOARCH)
}

func TestHandler_Environment(t *testing.T) {
	_, addr := setupHandler(t)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NilError(t, err)
	client := executorpb.NewExecutorClient(conn)
	envResp, err := client.Environment(context.Background(), &executorpb.EnvironmentRequest{})
	assert.NilError(t, err)
	assert.Assert(t, len(envResp.Environment) > 0)
}
