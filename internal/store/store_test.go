package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tidwall/buntdb"
)

func TestOpenAndClose(t *testing.T) {
	tmpDir := t.TempDir()

	s, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}

	if s.DB() == nil {
		t.Fatal("DB instance is nil")
	}

	if err := s.Close(); err != nil {
		t.Errorf("failed to close store: %v", err)
	}

	// Verify file was created
	dbPath := filepath.Join(tmpDir, dbFileName)
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("expected DB file to be created at %s", dbPath)
	}
}

func TestOpenInMemoryFallback(t *testing.T) {
	s, err := Open("/dev/null/invalid_dir")
	if err != nil {
		t.Fatalf("expected fallback to memory, got error: %v", err)
	}

	if s.DB() == nil {
		t.Fatal("DB instance is nil after fallback")
	}

	s.Close()
}

func TestCloseReleasesLockFile(t *testing.T) {
	tmpDir := t.TempDir()

	s, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	err = s.DB().Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set("k", "v", nil)
		return err
	})
	if err != nil {
		t.Fatalf("set: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("failed to close store: %v", err)
	}

	// mcp-buntdb v1.0.1 ReleaseDBLock closes the handle and leaves the file.
	// A held handle makes the next OpenBuntDB degrade to memory, and on
	// windows it makes t.TempDir RemoveAll fail (0011-MADR).
	lockPath := filepath.Join(tmpDir, dbFileName+".lock")
	if _, err := os.Stat(lockPath); err != nil {
		t.Fatalf("lock file should remain after Close: %v", err)
	}

	s2, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer s2.Close()
	var got string
	err = s2.DB().View(func(tx *buntdb.Tx) error {
		val, err := tx.Get("k")
		if err != nil {
			return err
		}
		got = val
		return nil
	})
	if err != nil || got != "v" {
		t.Fatalf("second open did not read the file db (held lock degrades to memory): got %q err %v", got, err)
	}
}
