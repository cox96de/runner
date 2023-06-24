package main

import (
	"net"
	"os"
	"strconv"

	"github.com/pkg/errors"

	"github.com/cox96de/runner/app/executor"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

type Config struct {
	Port       int
	SocketPath string
}

func main() {
	if err := increaseMaxOpenFiles(); err != nil {
		log.Warningf("failed to increase max open files: %v", err)
	}
	config, err := loadConfig(os.Args)
	checkError(err)
	listener, err := composeListener(config)
	checkError(err)
	app := executor.NewApp()
	if err := app.Run(listener); err != nil {
		log.Fatal(err)
	}
}

func loadConfig(arguments []string) (*Config, error) {
	flagSet := pflag.NewFlagSet("executor", pflag.ContinueOnError)
	c := &Config{}
	flagSet.IntVarP(&c.Port, "port", "", 8080, "port to listen on")
	flagSet.StringVarP(&c.SocketPath, "socket-path", "", "", "path to unix socket")
	if err := flagSet.Parse(arguments); err != nil {
		return nil, errors.WithStack(err)
	}
	return c, nil
}

func checkError(err error) {
	if err == nil {
		return
	}
	log.Fatal(err)
}

func composeListener(c *Config) (net.Listener, error) {
	switch {
	case c.SocketPath != "":
		listener, err := net.ListenUnix("unix", &net.UnixAddr{
			Name: c.SocketPath,
			Net:  "unix",
		})
		return listener, err
	case c.Port != 0:
		listener, err := net.Listen("tcp", ":"+strconv.Itoa(c.Port))
		return listener, err
	}
	return nil, errors.Errorf("no listener configured")
}
