package server

import (
	"testing"

	"github.com/cox96de/runner/app/server"

	"github.com/cox96de/runner/app/server/eventhook"
	"github.com/cox96de/runner/mock"
	"gotest.tools/v3/fs"
)

type MockServer struct {
	*server.App
}

func NewMockServer(t *testing.T) *MockServer {
	dbClient := mock.NewMockDBConn(t)
	redis := mock.NewMockRedisConn(t)
	app := server.NewApp(&server.Config{
		DB:                   dbClient,
		LogPersistentStorage: server.NewLocalLogPersistentStorage(fs.NewDir(t, "baseDir").Path()),
		LogCacheStorage:      redis,
		Locker:               server.NewRedisLocker(redis),
		EventHookSender:      eventhook.NewNopSender(),
	})
	return &MockServer{
		app,
	}
}
