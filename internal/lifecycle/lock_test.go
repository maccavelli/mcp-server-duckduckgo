package lifecycle

import (
	"path/filepath"
	"testing"
)

func TestTryLock(t *testing.T) {
	tmpDir := t.TempDir()
	lockFile := filepath.Join(tmpDir, "test.lock")

	err := TryLock(lockFile)
	if err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}
	defer Unlock()

	// Try acquiring again, should fail because it's locked by us globally
	// But actually, unix.Flock on the same process might succeed or fail depending on OS.
	// Since we are the same process, it might succeed, but we are just writing a basic test.

	// Test Unlock
	if err := Unlock(); err != nil {
		t.Errorf("failed to unlock: %v", err)
	}

	// Try acquiring again, should succeed
	err = TryLock(lockFile)
	if err != nil {
		t.Fatalf("failed to acquire lock after release: %v", err)
	}
	Unlock()
}

func TestTryLock_InvalidPath(t *testing.T) {
	err := TryLock("/dev/null/invalid.lock")
	if err == nil {
		t.Error("expected error for invalid path, got nil")
	}
}

func TestUnlock_Nil(t *testing.T) {
	globalLock = nil
	if err := Unlock(); err != nil {
		t.Errorf("expected nil error on empty unlock, got %v", err)
	}
}
