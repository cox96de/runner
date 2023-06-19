package main

import (
	"github.com/cox96de/runner/app/executor"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"os"
)

type Config struct {
	Port int
}

func main() {
	if err := increaseMaxOpenFiles(); err != nil {
		log.Warningf("failed to increase max open files: %v", err)
	}
	config, err := loadConfig(os.Args)
	checkError(err)
	app := executor.NewApp()
	if err := app.Run(config.Port); err != nil {
		log.Fatal(err)
	}
}

func loadConfig(arguments []string) (*Config, error) {
	flagSet := pflag.NewFlagSet("executor", pflag.ExitOnError)
	c := &Config{}
	flagSet.IntVarP(&c.Port, "port", "", 8080, "port to listen on")
	if err := flagSet.Parse(arguments); err != nil {
		return nil, err
	}
	return c, nil
}

func checkError(err error) {
	if err == nil {
		return
	}
	log.Fatal(err)
}
