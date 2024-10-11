package main

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/cox96de/runner/githubapp/ghclient"

	"github.com/cox96de/runner/log"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/api/httpserverclient"
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
	runnerClient, err := httpserverclient.NewClient(&http.Client{}, config.RunnerURL)
	checkError(err)
	dbCli, err := ComposeDB(config.DB)
	checkError(err)
	ghClient := ghclient.NewClient(client)
	app := app.NewApp(ghClient, runnerClient, config.ExportURL, dbCli, config.CloneStep)
	dispatcher := githubapp.NewEventDispatcher([]githubapp.EventHandler{app}, "")
	engine := gin.New()
	group := engine.Group(config.BaseURL)
	group.Any("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%s", time.Now())})
	})
	group.POST("/webhook", gin.WrapH(dispatcher))
	group.GET("/log", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", logWebContent)
	})
	eventHandler, err := app.GetRunnerHandler(context.Background())
	checkError(err)
	group.POST("/runner_event", eventHandler)
	// FIXME: this api is not authenticated.
	group.GET("/job_executions/:job_execution_id/logs/:log_name", app.GetLogHandler)
	err = http.ListenAndServe(config.ListenAddr, engine)
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func ComposeDB(c *DB) (*db.Client, error) {
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
	client := db.NewClient(db.Dialect(dialect), conn)
	if db.Dialect(dialect) == db.SQLite {
		log.Warningf("sqlite database is not recommended for production use")
		if err := conn.AutoMigrate(&db.Pipeline{}, &db.Job{}); err != nil {
			return nil, err
		}
	}
	return client, nil
}
