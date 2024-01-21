package handler

import (
	"net"
	"testing"

	"github.com/cox96de/runner/app/executor/executorpb"
	"google.golang.org/grpc"

	"gotest.tools/v3/assert"
)

func setupHandler(t *testing.T) (*Handler, string) {
	handler := NewHandler()
	testServer := grpc.NewServer()
	executorpb.RegisterExecutorServer(testServer, handler)
	t.Cleanup(func() {
		testServer.Stop()
	})
	listener, err := net.Listen("tcp", ":0")
	assert.NilError(t, err)
	go func() {
		_ = testServer.Serve(listener)
	}()
	return handler, listener.Addr().String()
}
