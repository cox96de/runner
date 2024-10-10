//go:build linux || darwin

package handler

import (
	"os/user"
	"strconv"
	"syscall"

	"github.com/cockroachdb/errors"
)

func setUserForSysProcAttr(attr *syscall.SysProcAttr, u *user.User) error {
	uid, err := strconv.ParseInt(u.Uid, 10, 64)
	if err != nil {
		return errors.WithMessage(err, "failed to parse uid")
	}
	gid, err := strconv.ParseInt(u.Gid, 10, 64)
	if err != nil {
		return errors.WithMessage(err, "failed to parse gid")
	}
	attr.Credential = &syscall.Credential{
		Uid: uint32(uid),
		Gid: uint32(gid),
	}
	return nil
}
