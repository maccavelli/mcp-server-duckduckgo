package store

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/tidwall/buntdb"
)

const (
	// healthPrefix is the BuntDB key prefix for provider health records.
	healthPrefix = "health:"

	// HealthTTL is the time-to-live for health tracking entries.
	HealthTTL = 1 * time.Hour

	// MaxConsecFails is the threshold after which a provider is skipped.
	MaxConsecFails = 3
)

// ProviderHealth tracks per-provider success/failure/CAPTCHA metrics.
type ProviderHealth struct {
	LastSuccess  time.Time `json:"last_ok"`
	LastFailure  time.Time `json:"last_fail"`
	ConsecFails  int       `json:"consec_fails"`
	CaptchaCount int       `json:"captcha_count"`
}

// RecordSuccess resets the consecutive failure counter and updates the
// last success timestamp for the given provider.
func (s *Store) RecordSuccess(provider string) {
	s.updateHealth(provider, func(h *ProviderHealth) {
		h.LastSuccess = time.Now()
		h.ConsecFails = 0
	})
}

// RecordFailure increments the consecutive failure counter and updates
// the last failure timestamp for the given provider.
func (s *Store) RecordFailure(provider string) {
	s.updateHealth(provider, func(h *ProviderHealth) {
		h.LastFailure = time.Now()
		h.ConsecFails++
	})
}

// RecordCAPTCHA increments the CAPTCHA counter and also records the
// event as a failure for the given provider.
func (s *Store) RecordCAPTCHA(provider string) {
	s.updateHealth(provider, func(h *ProviderHealth) {
		h.LastFailure = time.Now()
		h.ConsecFails++
		h.CaptchaCount++
	})
}

// GetHealth retrieves the current health record for a provider.
func (s *Store) GetHealth(provider string) ProviderHealth {
	var h ProviderHealth
	if err := s.db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(healthPrefix + provider)
		if err != nil {
			return err
		}
		return json.Unmarshal([]byte(val), &h)
	}); err != nil {
		slog.Debug("failed to read provider health", "provider", provider, "error", err)
	}
	return h
}

// IsHealthy returns true if the provider has fewer than MaxConsecFails
// consecutive failures in the current health window.
func (s *Store) IsHealthy(provider string) bool {
	h := s.GetHealth(provider)
	return h.ConsecFails < MaxConsecFails
}

// updateHealth applies a mutation function to a provider's health record
// inside an atomic BuntDB transaction.
func (s *Store) updateHealth(provider string, mutate func(*ProviderHealth)) {
	key := healthPrefix + provider
	if err := s.db.Update(func(tx *buntdb.Tx) error {
		var h ProviderHealth
		val, err := tx.Get(key)
		if err == nil {
			if unmarshalErr := json.Unmarshal([]byte(val), &h); unmarshalErr != nil {
				slog.Debug("failed to unmarshal provider health", "provider", provider, "error", unmarshalErr)
			}
		}

		mutate(&h)

		data, marshalErr := json.Marshal(h)
		if marshalErr != nil {
			return marshalErr
		}
		_, _, err = tx.Set(key, string(data), &buntdb.SetOptions{
			Expires: true,
			TTL:     HealthTTL,
		})
		return err
	}); err != nil {
		slog.Warn("failed to update provider health", "provider", provider, "error", err)
	}
}
