package store

import (
	"testing"
	"time"
)

func TestVQDOperations(t *testing.T) {
	tmpDir := t.TempDir()
	s, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer s.Close()

	query := "test query"
	vqdValue := "vqd_token_123"

	// Should not exist initially
	_, ok := s.GetVQD(query)
	if ok {
		t.Error("expected VQD to not exist initially")
	}

	// Set VQD
	s.SetVQD(query, vqdValue, VQDDefaultTTL)

	// Get VQD
	val, ok := s.GetVQD(query)
	if !ok {
		t.Error("expected VQD to exist")
	}
	if val != vqdValue {
		t.Errorf("expected %s, got %s", vqdValue, val)
	}

	// Set with very short TTL and wait
	s.SetVQD("short", "val", 1*time.Millisecond)
	time.Sleep(10 * time.Millisecond)

	_, ok = s.GetVQD("short")
	if ok {
		t.Error("expected VQD to expire")
	}
}
