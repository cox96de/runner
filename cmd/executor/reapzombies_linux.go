package main

import (
	"context"
	"os"
	"runtime"
	"time"

	"github.com/cox96de/runner/cmd/executor/zombies"
	"github.com/cox96de/runner/log"
)

func reapZombies() {
	if runtime.GOOS != "linux" {
		return
	}
	if os.Getpid() != 1 {
		log.Infof("pid is not 1, no need to reap zombie processes")
		return
	}
	go func() {
		if err := zombies.RunReap(context.Background(), time.Second); err != nil {
			log.Errorf("failed to reap zombies process: %+v", err)
		}
		log.Infof("the worker to reap zombies is exited")
	}()
}
