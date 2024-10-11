package handler

import (
	"os/user"
	"syscall"
)

func setUserForSysProcAttr(attr *syscall.SysProcAttr, u *user.User) error {
	// Windows doesn't have the concept of setting the user for a process.
	return nil
}
