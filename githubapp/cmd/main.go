package main

import (
	"context"
	_ "embed"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/app/server"
	"github.com/redis/go-redis/extra/redisotel/v9"
	goredis "github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/cox96de/runner/githubapp/ghclient"

	"github.com/cox96de/runner/log"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/githubapp/app"
	"github.com/cox96de/runner/githubapp/db"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v64/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/spf13/pflag"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gopkg.in/yaml.v2"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

//go:embed log.html
var logWebContent []byte

func main() {
	var configFilePath string
	flags := pflag.NewFlagSet("githubapp", pflag.ContinueOnError)
	flags.StringVarP(&configFilePath, "config", "c", "config.yaml", "path to config file")
	err := flags.Parse(os.Args[1:])
	checkError(err)
	file, err := os.ReadFile(configFilePath)
	checkError(err)
	var config Config
	err = yaml.UnmarshalStrict(file, &config)
	checkError(err)
	transport, err := ghinstallation.NewAppsTransport(http.DefaultTransport, config.GithubAppID, []byte(config.PrivateKey))
	checkError(err)
	client := github.NewClient(&http.Client{Transport: transport})
	dbConn, err := ComposeDB(config.DB)
	checkError(err)
	ghClient := ghclient.NewClient(client)
	redis, err := ComposeRedis(config.RunnerServer.Redis)
	checkError(err)
	var a *app.App
	runnerDB, err := ComposeDB(config.RunnerServer.DB)
	checkError(err)
	dbCli := db.NewClient(db.Dialect(dbConn.Dialector.Name()), dbConn)
	a = app.NewApp(ghClient, config.ExportURL, dbCli, config.CloneStep)
	runnerServer := server.NewApp(&server.Config{
		DB:                   runnerDB,
		LogPersistentStorage: server.NewLocalLogPersistentStorage(config.RunnerServer.LogArchiveDir),
		LogCacheStorage:      redis,
		Locker:               server.NewRedisLocker(redis),
		EventHookSender:      a,
	})
	a.SetRunnerServer(runnerServer)
	dispatcher := githubapp.NewEventDispatcher([]githubapp.EventHandler{a}, "")
	engine := gin.New()
	group := engine.Group(config.BaseURL)
	group.Any("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%s", time.Now())})
	})
	group.POST("/webhook", gin.WrapH(dispatcher))
	group.GET("/log", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", logWebContent)
	})
	eventHandler, err := a.GetRunnerHandler(context.Background())
	checkError(err)
	group.POST("/runner_event", eventHandler)
	group.POST("/runner_event/:job_execution_id", a.HandleJobExecutionRefresh)
	// FIXME: this api is not authenticated.
	group.GET("/job_executions/:job_execution_id/logs/:log_name", a.GetLogHandler)
	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	api.RegisterServerServer(grpcServer, runnerServer)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.RunnerServer.GRPCPort))
	checkError(err)
	log.Infof("listenning and serving grpc on %s", listener.Addr().String())
	g := &errgroup.Group{}
	g.Go(func() error {
		return grpcServer.Serve(listener)
	})
	g.Go(func() error {
		return http.ListenAndServe(config.ListenAddr, engine)
	})
	err = g.Wait()
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func ComposeDB(c *DB) (*gorm.DB, error) {
	var (
		conn    *gorm.DB
		err     error
		dialect = c.Dialect
		dsn     = c.DSN
	)
	opts := &gorm.Config{}
	if c.TablePrefix != "" {
		opts.NamingStrategy = &schema.NamingStrategy{
			TablePrefix: c.TablePrefix,
		}
	}
	switch db.Dialect(dialect) {
	case db.Mysql:
		conn, err = gorm.Open(mysql.Open(dsn), opts)
	case db.Postgres:
		conn, err = gorm.Open(postgres.Open(dsn), opts)
	case db.SQLite:
		conn, err = gorm.Open(sqlite.Open(dsn), opts)
	default:
		return nil, errors.Errorf("unsupported dialect: %s", dialect)
	}
	if err != nil {
		return nil, errors.WithMessage(err, "failed to open database connection")
	}
	if err = conn.Use(otelgorm.NewPlugin()); err != nil {
		return nil, errors.WithMessage(err, "failed to use otelgorm plugin")
	}
	return conn, nil
}

func ComposeRedis(r *Redis) (*goredis.Client, error) {
	if r.Addr == "internal" {
		return ComposeInternalRedis()
	}
	conn := goredis.NewClient(&goredis.Options{
		Addr:                  r.Addr,
		Username:              r.Username,
		Password:              r.Password,
		DB:                    r.DB,
		MaxRetries:            r.MaxRetries,
		MinRetryBackoff:       r.MinRetryBackoff,
		MaxRetryBackoff:       r.MaxRetryBackoff,
		DialTimeout:           r.DialTimeout,
		ReadTimeout:           r.ReadTimeout,
		WriteTimeout:          r.WriteTimeout,
		ContextTimeoutEnabled: true,
		PoolFIFO:              r.PoolFIFO,
		PoolSize:              r.PoolSize,
		PoolTimeout:           r.PoolTimeout,
		MinIdleConns:          r.MinIdleConns,
		MaxIdleConns:          r.MaxIdleConns,
		MaxActiveConns:        r.MaxActiveConns,
		ConnMaxIdleTime:       r.ConnMaxIdleTime,
		ConnMaxLifetime:       r.ConnMaxLifetime,
	})
	if err := redisotel.InstrumentTracing(conn); err != nil {
		return nil, errors.WithMessage(err, "failed to instrument tracing")
	}
	if err := redisotel.InstrumentMetrics(conn); err != nil {
		return nil, errors.WithMessage(err, "failed to instrument metrics")
	}
	return conn, nil
}

// ComposeInternalRedis returns a redis client with internal redis.
func ComposeInternalRedis() (*goredis.Client, error) {
	log.Warningf("You are using internal redis, it is not recommended for production.")
	miniRedis := miniredis.NewMiniRedis()
	if err := miniRedis.Start(); err != nil {
		return nil, err
	}
	conn := goredis.NewClient(&goredis.Options{
		Network: "",
		Addr:    miniRedis.Addr(),
	})
	if err := redisotel.InstrumentTracing(conn); err != nil {
		return nil, errors.WithMessage(err, "failed to instrument tracing")
	}
	if err := redisotel.InstrumentMetrics(conn); err != nil {
		return nil, errors.WithMessage(err, "failed to instrument metrics")
	}
	return conn, nil
}
