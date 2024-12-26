//go:build linux || darwin

package handler

import (
	"os/user"
	"strconv"
	"syscall"

	"github.com/cox96de/runner/log"

	"github.com/cockroachdb/errors"
)

func init() {
	var err error
	realUID, effectiveUID := syscall.Getuid(), syscall.Geteuid()
	realUser, err = user.LookupId(strconv.Itoa(realUID))
	if err != nil {
		log.Fatalf("failed to find real user by uid %d: %v", realUID, err)
	}
	effectiveUser, err = user.LookupId(strconv.Itoa(effectiveUID))
	if err != nil {
		log.Fatalf("failed to find effective user by uid %d: %v", effectiveUID, err)
	}
}

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
