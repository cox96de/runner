package main

import (
	"fmt"

	"github.com/cox96de/runner/app/server/dispatch"
	"github.com/cox96de/runner/app/server/pipeline"

	"github.com/cox96de/runner/app/server/handler"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	command := GetServerCommand()
	root := &cobra.Command{}
	root.AddCommand(command)
	err := root.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func GetServerCommand() *cobra.Command {
	var configPath string
	c := &cobra.Command{
		Use: "server",
		Run: func(cmd *cobra.Command, args []string) {
			err := RunServer(configPath)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
	flags := c.Flags()
	flags.StringVarP(&configPath, "config", "c", "config.yaml", "path to config file")
	return c
}

func RunServer(configfile string) error {
	config, err := LoadConfig(configfile)
	if err != nil {
		return errors.WithMessage(err, "failed to load config")
	}
	dbClient, err := ComposeDB(config.DB.Dialect, config.DB.DSN)
	if err != nil {
		return errors.WithMessage(err, "failed to compose db")
	}
	locker, err := ComposeLocker(config.Locker)
	if err != nil {
		return errors.WithMessage(err, "failed to compose locker")
	}
	h := handler.NewHandler(dbClient, pipeline.NewService(dbClient), dispatch.NewService(dbClient), locker)
	engine := gin.New()
	group := engine.Group("/api/v1")
	h.RegisterRouter(group)
	return engine.Run(fmt.Sprintf(":%d", config.Port))
}
