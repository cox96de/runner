package handler

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"

	"github.com/cockroachdb/errors"
)

type User struct {
	*user.User
	// PosixUid is the user's uid, available on POSIX systems.
	PosixUid int
	// PosixGid is the user's gid, available on POSIX systems.
	PosixGid int
}

func lookupUser(username string) (*User, error) {
	var (
		u   *user.User
		err error
	)
	u, err = user.Lookup(username)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to lookup user")
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to parse uid")
	}
	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to parse gid")
	}
	return &User{
		User:     u,
		PosixUid: uid,
		PosixGid: gid,
	}, nil
}

func mkdirAll(path string, perm os.FileMode, username string) error {
	// Fast path: if we can tell whether path is a directory or file, stop with success or error.
	dir, err := os.Stat(path)
	if err == nil {
		if dir.IsDir() {
			return nil
		}
		return &os.PathError{Op: "mkdir", Path: path, Err: syscall.ENOTDIR}
	}
	var u *User
	if len(username) > 0 {
		u, err = lookupUser(username)
		if err != nil {
			return errors.WithMessage(err, "failed to lookup user")
		}
	}
	return mkdirAll2(path, perm, u)
}

func mkdirAll2(path string, perm os.FileMode, username *User) error {
	var err error
	// Slow path: make sure parent exists and then call Mkdir for path.
	i := len(path)
	for i > 0 && os.IsPathSeparator(path[i-1]) { // Skip trailing path separator.
		i--
	}

	j := i
	for j > 0 && !os.IsPathSeparator(path[j-1]) { // Scan backward over element.
		j--
	}

	if j > 1 {
		// Create parent.
		err = mkdirAll2(fixRootDirectory(path[:j-1]), perm, username)
		if err != nil {
			return err
		}
	}

	// Parent now exists; invoke Mkdir and use its result.
	err = os.Mkdir(path, perm)
	if err != nil {
		// Handle arguments like "foo/." by
		// double-checking that directory doesn't exist.
		dir, err1 := os.Lstat(path)
		if err1 == nil {
			if dir.IsDir() {
				return nil
			}
			if dir.Mode()|os.ModeSymlink != 0 {
				target, err := filepath.EvalSymlinks(path)
				if err != nil {
					return err
				}
				dir, err := os.Lstat(target)
				if err != nil {
					return err
				}
				if dir.IsDir() {
					return nil
				}
			}
		}
		return err
	}
	_ = os.Chmod(path, perm)
	if username != nil {
		if err := os.Chown(path, username.PosixUid, username.PosixGid); err != nil {
			return errors.WithMessagef(err, "failed to chown for new created path %s", path)
		}
	}
	return nil
}

func fixRootDirectory(p string) string {
	if runtime.GOOS == "windows" && len(p) == len(`\\?\c:`) &&
		os.IsPathSeparator(p[0]) && os.IsPathSeparator(p[1]) && p[2] == '?' &&
		os.IsPathSeparator(p[3]) && p[5] == ':' {
		return p + `\`
	}
	return p
}
