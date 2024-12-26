package handler

import (
	"os/user"
	"syscall"

	"github.com/cox96de/runner/log"
)

func init() {
	var err error
	realUser, err = user.Current()
	if err != nil {
		log.Fatalf("failed to get current user: %v", err)
	}
	effectiveUser = realUser
}

func setUserForSysProcAttr(_ *syscall.SysProcAttr, _ *user.User) error {
	// Windows doesn't have the concept of setting the user for a process.
	return nil
}
