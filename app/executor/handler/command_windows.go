package handler

import (
	"os/user"
	"syscall"
)

func init() {
	var err error
	realUser, err = user.Current()
	if err != nil {
		log.Fatalf("failed to find real user by uid %d: %v", realUID, err)
	}
	effectiveUser = realUser
}

func setUserForSysProcAttr(attr *syscall.SysProcAttr, u *user.User) error {
	// Windows doesn't have the concept of setting the user for a process.
	return nil
}
