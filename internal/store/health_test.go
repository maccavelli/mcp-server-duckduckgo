package store

import (
	"testing"
)

func TestHealthTracking(t *testing.T) {
	tmpDir := t.TempDir()
	s, err := Open(tmpDir)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer s.Close()

	provider := "duckduckgo"

	// Initially healthy
	if !s.IsHealthy(provider) {
		t.Error("expected provider to be healthy initially")
	}

	// Record failures
	s.RecordFailure(provider)
	s.RecordFailure(provider)

	h := s.GetHealth(provider)
	if h.ConsecFails != 2 {
		t.Errorf("expected 2 consecutive failures, got %d", h.ConsecFails)
	}

	// Record CAPTCHA
	s.RecordCAPTCHA(provider)
	h = s.GetHealth(provider)
	if h.ConsecFails != 3 {
		t.Errorf("expected 3 consecutive failures, got %d", h.ConsecFails)
	}
	if h.CaptchaCount != 1 {
		t.Errorf("expected 1 captcha, got %d", h.CaptchaCount)
	}

	// Should now be unhealthy
	if s.IsHealthy(provider) {
		t.Error("expected provider to be unhealthy")
	}

	// Record success resets fails
	s.RecordSuccess(provider)
	h = s.GetHealth(provider)
	if h.ConsecFails != 0 {
		t.Errorf("expected 0 consecutive failures after success, got %d", h.ConsecFails)
	}
	if !s.IsHealthy(provider) {
		t.Error("expected provider to be healthy after success")
	}
}
