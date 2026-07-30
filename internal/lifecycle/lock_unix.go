//go:build !windows

package lifecycle

import (
	"os"
	"path/filepath"
	"strconv"

	"golang.org/x/sys/unix"
)

var globalLock *os.File

func TryLock(lockPath string) error {
	cleanPath := filepath.Clean(lockPath)
	//nolint:gosec // G304: lock path is constructed from the user cache directory
	f, err := os.OpenFile(cleanPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}

	err = unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
	if err != nil {
		if closeErr := f.Close(); closeErr != nil {
			return closeErr
		}
		return err
	}

	if err := f.Truncate(0); err != nil {
		return err
	}
	if _, err := f.WriteString(strconv.Itoa(os.Getpid())); err != nil {
		return err
	}

	globalLock = f
	return nil
}

func Unlock() error {
	if globalLock != nil {
		if err := unix.Flock(int(globalLock.Fd()), unix.LOCK_UN); err != nil {
			return err
		}
		err := globalLock.Close()
		globalLock = nil
		return err
	}
	return nil
}
