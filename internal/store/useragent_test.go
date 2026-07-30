package store

import (
	"testing"
)

func TestUserAgentRotation(t *testing.T) {
	tmpDir := t.TempDir()
	s, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer s.Close()

	provider := "duckduckgo"

	// Seed pool
	s.SeedUAPool()

	// Get initial UA
	ua1 := s.GetUA(provider)
	if ua1 == "" {
		t.Error("expected non-empty user agent")
	}

	// Getting again should return the same UA due to rotation TTL
	ua2 := s.GetUA(provider)
	if ua1 != ua2 {
		t.Errorf("expected same user agent, got %s and %s", ua1, ua2)
	}

	// Blacklist current UA
	s.BlacklistUA(ua1)

	// Get UA should return a new one
	ua3 := s.GetUA(provider)
	if ua3 == ua1 {
		t.Error("expected new user agent after blacklisting")
	}
}
