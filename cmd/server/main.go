package main

import (
	"fmt"
	"time"

	"github.com/andeya/goutil/calendar/cron"
	"github.com/cox96de/runner/app/server/monitor"
	"golang.org/x/net/context"

	"github.com/cox96de/runner/db"
	"github.com/cox96de/runner/util"
	"github.com/spf13/viper"

	"github.com/cox96de/runner/app/server/dispatch"
	"github.com/cox96de/runner/app/server/pipeline"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/app/server/handler"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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
			err = RunServer(&config)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
	flags := c.Flags()
	flags.StringVarP(&configFilePath, "config", "c", "config.yaml", "path to config file")
	checkError(util.BindIntArg(flags, vv, &util.IntArg{
		ArgKey:    "port",
		FlagName:  "port",
		FlagValue: 8080,
		FlagUsage: "port to listen",
		Env:       "RUNNER_PORT",
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
	h := handler.NewHandler(dbClient, pipeline.NewService(dbClient), dispatch.NewService(dbClient), locker, logStorage)
	engine := gin.New()
	group := engine.Group("/api/v1")
	h.RegisterRouter(group)
	startConJob(context.Background(), monitor.NewService(dbClient))
	return engine.Run(fmt.Sprintf(":%d", config.Port))
}

func startConJob(ctx context.Context, service *monitor.Service) {
	c := cron.New()
	err := c.AddFunc("@every 1m", func() {
		if err := service.RecycleHeartbeatTimeoutJobs(ctx, time.Minute); err != nil {
			log.Errorf("failed to rcycle heartbeat timeout job executions")
		}
	})
	checkError(err)
	c.Start()
}
