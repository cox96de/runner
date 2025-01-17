package server

import (
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/cox96de/runner/app/server/dispatch"
	"github.com/cox96de/runner/app/server/eventhook"
	"github.com/cox96de/runner/app/server/handler"
	"github.com/cox96de/runner/app/server/logstorage"
	"github.com/cox96de/runner/app/server/monitor"
	"github.com/cox96de/runner/app/server/pipeline"
	"github.com/cox96de/runner/db"
	"github.com/cox96de/runner/external/redis"
	"github.com/cox96de/runner/lib"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type (
	LogPersistentStorage logstorage.OSS
	Locker               lib.Locker
	EventHookSender      eventhook.Sender
)

type App struct {
	*handler.Handler
	*monitor.Service
}

type Config struct {
	DB                   *gorm.DB
	LogPersistentStorage LogPersistentStorage
	LogCacheStorage      *goredis.Client
	Locker               Locker
	EventHookSender      EventHookSender
}

func NewApp(c *Config) *App {
	logStorage := composeLogStorage(c.LogCacheStorage, c.LogPersistentStorage)
	dbClient := db.NewClient(c.DB)
	eventhookService := eventhook.NewService(c.EventHookSender)
	dispatchService := dispatch.NewService(dbClient, eventhookService)
	h := handler.NewHandler(dbClient, pipeline.NewService(dbClient), dispatchService, c.Locker, logStorage,
		eventhookService)
	monitorService := monitor.NewService(dbClient, logStorage, eventhookService, dispatchService)
	return &App{
		Handler: h,
		Service: monitorService,
	}
}

func composeLogStorage(cache *goredis.Client, persistent LogPersistentStorage) *logstorage.Service {
	composeRedis := redis.NewClient(cache)
	return logstorage.NewService(composeRedis, persistent)
}

func NewRedisLocker(client *goredis.Client) Locker {
	return redis.NewClient(client)
}

func NewLocalLogPersistentStorage(dir string) LogPersistentStorage {
	return logstorage.NewFilesystemOSS(dir)
}

func NewS3LogPersistentStorage(bucket string, s3 s3iface.S3API) LogPersistentStorage {
	return logstorage.NewS3(bucket, s3)
}
