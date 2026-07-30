//go:build windows

package lifecycle

import (
	"os"
	"strconv"

	"golang.org/x/sys/windows"
)

var globalLock *os.File

func TryLock(lockPath string) error {
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	var overlapped windows.Overlapped
	err = windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &overlapped)
	if err != nil {
		f.Close()
		return err
	}

	f.Truncate(0)
	f.WriteString(strconv.Itoa(os.Getpid()))

	globalLock = f
	return nil
}

func Unlock() error {
	if globalLock != nil {
		var overlapped windows.Overlapped
		windows.UnlockFileEx(windows.Handle(globalLock.Fd()), 0, 1, 0, &overlapped)
		err := globalLock.Close()
		globalLock = nil
		return err
	}
	return nil
}
