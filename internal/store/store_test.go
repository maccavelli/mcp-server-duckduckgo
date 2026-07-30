package store

import (
	"os"
	"path/filepath"
	"testing"
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
