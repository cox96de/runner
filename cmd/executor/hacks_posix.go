//go:build linux || darwin

package main

import "syscall"

func increaseMaxOpenFiles() error {
	var limit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit)
	if err == nil {
		limit.Cur = limit.Max
		err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &limit)
	}
	return err
}
