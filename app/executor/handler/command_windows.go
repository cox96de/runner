package handler

import (
	"os/user"
	"syscall"
)

func setUserForSysProcAttr(attr *syscall.SysProcAttr, u *user.User) error {
	return nil
}
