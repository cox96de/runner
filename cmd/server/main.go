package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/sync/errgroup"

	"github.com/cox96de/runner/api"
	"google.golang.org/grpc"

	"github.com/cox96de/runner/app/server/eventhook"

	"github.com/cox96de/runner/log"
	"github.com/cox96de/runner/telemetry/trace"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/andeya/goutil/calendar/cron"
	"github.com/cox96de/runner/app/server/monitor"
	"github.com/cox96de/runner/db"
	"github.com/cox96de/runner/util"
	"github.com/spf13/viper"

	"github.com/cox96de/runner/app/server/dispatch"
	"github.com/cox96de/runner/app/server/pipeline"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/app/server/handler"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

func main() {
	command := GetServerCommand()
	err := command.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func GetServerCommand() *cobra.Command {
	vv := viper.New()
	var configFilePath string
	c := &cobra.Command{
		Use: "server",
		Run: func(cmd *cobra.Command, args []string) {
			if len(configFilePath) > 0 {
				vv.SetConfigFile(configFilePath)
			}
			var config Config
			err := vv.UnmarshalExact(&config)
			if err != nil {
				log.Fatalf("failed to load config: %v", err)
			}
			log.SetLevel(log.DebugLevel)
			shutdown, err := setupOTEL(cmd.Context())
			checkError(err)
			defer func() {
				_ = shutdown(cmd.Context())
			}()
			err = RunServer(&config)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
	flags := c.Flags()
	flags.StringVarP(&configFilePath, "config", "c", "config.yaml", "path to config file")
	checkError(util.BindIntArg(flags, vv, &util.IntArg{
		ArgKey:    "http.port",
		FlagName:  "http.port",
		FlagValue: 8080,
		FlagUsage: "port to http listen",
		Env:       "RUNNER_HTTP_PORT",
	}))
	checkError(util.BindIntArg(flags, vv, &util.IntArg{
		ArgKey:    "grpc.port",
		FlagName:  "grpc.port",
		FlagValue: 7080,
		FlagUsage: "port to grpc listen",
		Env:       "RUNNER_GRPC_PORT",
	}))
	checkError(util.BindStringArg(flags, vv, &util.StringArg{
		ArgKey:    "db.dialect",
		FlagName:  "db.dialect",
		FlagValue: string(db.SQLite),
		FlagUsage: "db dialect type, support sqlite, mysql, postgres",
		Env:       "RUNNER_DB_DIALECT",
	}))
	checkError(util.BindStringArg(flags, vv, &util.StringArg{
		ArgKey:    "db.dsn",
		FlagName:  "db.dsn",
		FlagValue: "file:/data/db/sqlite3.db?cache=shared",
		FlagUsage: "db dsn",
		Env:       "RUNNER_DB_DSN",
	}))
	checkError(util.BindStringArg(flags, vv, &util.StringArg{
		ArgKey:    "locker.backend",
		FlagName:  "locker.backend",
		FlagValue: "redis",
		FlagUsage: "the distribute lock type, support redis",
		Env:       "RUNNER_LOCKER_TYPE",
	}))
	checkError(util.BindStringArg(flags, vv, &util.StringArg{
		ArgKey:    "locker.redis.addr",
		FlagName:  "locker.redis.addr",
		FlagValue: "internal",
		FlagUsage: "redis address",
		Env:       "RUNNER_LOCKER_REDIS_ADDR",
	}))
	checkError(util.BindStringArg(flags, vv, &util.StringArg{
		ArgKey:    "log_storage.redis.addr",
		FlagName:  "log_storage.redis.addr",
		FlagValue: "internal",
		FlagUsage: "the redis address for log storage, it is used for temporary storage",
		Env:       "RUNNER_LOG_STORAGE_REDIS_ADDR",
	}))
	checkError(util.BindStringArg(flags, vv, &util.StringArg{
		ArgKey:    "log_storage.archive.backend",
		FlagName:  "log_storage.archive.backend",
		FlagValue: "fs",
		FlagUsage: "the backend of building log persistence storage, support fs",
		Env:       "RUNNER_LOG_STORAGE_ARCHIVE_BACKEND",
	}))
	checkError(util.BindStringArg(flags, vv, &util.StringArg{
		ArgKey:    "log_storage.archive.base_dir",
		FlagName:  "log_storage.archive.base_dir",
		FlagValue: "/data/logs",
		FlagUsage: "the base dir of log archive. For fs backend, it is the root dir of log storage",
		Env:       "RUNNER_LOG_STORAGE_ARCHIVE_BASE_DIR",
	}))
	checkError(util.BindStringArg(flags, vv, &util.StringArg{
		ArgKey:    "event.http_endpoint",
		FlagName:  "event.http_endpoint",
		FlagValue: "",
		FlagUsage: "enable http event endpoint",
		Env:       "RUNNER_EVENT_HTTP_ENDPOINT",
	}))
	return c
}

func checkError(err error) {
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func RunServer(config *Config) error {
	dbClient, err := ComposeDB(config.DB)
	if err != nil {
		return errors.WithMessage(err, "failed to compose db")
	}
	locker, err := ComposeLocker(config.Locker)
	if err != nil {
		return errors.WithMessage(err, "failed to compose locker")
	}
	logStorage, err := ComposeLogStorage(config.LogStorage)
	if err != nil {
		return errors.WithMessage(err, "failed to compose log storage")
	}
	cloudEventsClient, err := ComposeCloudEventsClient(config.Event)
	if err != nil {
		return errors.WithMessage(err, "failed to compose cloud events client")
	}
	eventhookService := eventhook.NewService(cloudEventsClient)
	dispatchService := dispatch.NewService(dbClient, eventhookService)
	h := handler.NewHandler(dbClient, pipeline.NewService(dbClient), dispatchService, locker, logStorage,
		eventhookService)
	errgroup := errgroup.Group{}
	if config.HTTP.Port > 0 {
		errgroup.Go(func() error {
			engine := gin.New()
			// It's important to set the context with fallback to true, so that the context will be propagated to the next middleware.
			engine.ContextWithFallback = true
			engine.Use(otelgin.Middleware("runner-server"))
			group := engine.Group("/api/v1")
			h.RegisterRouter(group)
			startConJob(monitor.NewService(dbClient, logStorage, eventhookService, dispatchService))
			return engine.Run(fmt.Sprintf(":%d", config.HTTP.Port))
		})
	}
	if config.GRPC.Port > 0 {
		errgroup.Go(func() error {
			server := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
			api.RegisterServerServer(server, h)
			listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.GRPC.Port))
			checkError(err)
			log.Infof("listenning and serving grpc on %s", listener.Addr().String())
			return server.Serve(listener)
		})
	}
	return errgroup.Wait()
}

func startConJob(service *monitor.Service) {
	c := cron.New()
	err := c.AddFunc("@every 1m", func() {
		ctx, span := trace.Start(context.Background(), "cronjob.recycle_heartbeat_timeout_jobs")
		defer span.End()
		if err := service.RecycleHeartbeatTimeoutJobs(ctx, time.Minute); err != nil {
			log.Errorf("failed to rcycle heartbeat timeout job executions")
		}
	})
	checkError(err)
	c.Start()
}
