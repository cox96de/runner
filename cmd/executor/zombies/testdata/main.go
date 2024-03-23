package main

import (
	"context"
	"github.com/cox96de/runner/cmd/executor/zombies"
	"github.com/cox96de/runner/log"
	"os"
	"os/exec"
	"time"
)

func main() {
	run("pip", "install", "psutil")
	go func() {
		err := zombies.RunReap(context.Background(), time.Second)
		if err != nil {
			panic(err)
		}
		log.Info("reap zombies stopped")
	}()
	run(os.Args[1])
}

func run(cmds ...string) {
	command := exec.Command(cmds[0], cmds[1:]...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err := command.Run()
	if err != nil {
		panic(err)
	}
}
