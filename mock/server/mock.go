package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cox96de/runner/api/httpserverclient"
	"github.com/cox96de/runner/app/server/dispatch"
	"github.com/cox96de/runner/app/server/eventhook"
	"github.com/cox96de/runner/app/server/handler"
	"github.com/cox96de/runner/app/server/logstorage"
	"github.com/cox96de/runner/app/server/pipeline"
	"github.com/cox96de/runner/mock"
	"github.com/gin-gonic/gin"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"
)

type MockServer struct {
	*httpserverclient.Client
}

func NewMockServer(t *testing.T) *MockServer {
	dbClient := mock.NewMockDB(t)
	eventhook := eventhook.NewService(eventhook.NewNopSender())
	h := handler.NewHandler(dbClient, pipeline.NewService(dbClient), dispatch.NewService(dbClient, eventhook), mock.NewMockLocker(),
		logstorage.NewService(mock.NewMockRedis(t), logstorage.NewFilesystemOSS(fs.NewDir(t, "baseDir").Path())), eventhook)
	engine := gin.New()
	h.RegisterRouter(engine.Group("/api/v1"))
	server := httptest.NewServer(engine)
	t.Cleanup(func() {
		server.Close()
	})
	client, err := httpserverclient.NewClient(&http.Client{}, server.URL)
	assert.NilError(t, err)
	return &MockServer{
		Client: client,
	}
}
