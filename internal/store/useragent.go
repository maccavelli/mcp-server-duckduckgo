package store

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/tidwall/buntdb"
)

const (
	// uaPoolKey stores the JSON-encoded UA pool.
	uaPoolKey = "ua:pool"

	// uaCurrentPrefix stores the active UA per provider.
	uaCurrentPrefix = "ua:current:"

	// uaBlacklistPrefix stores temporarily blacklisted UAs.
	uaBlacklistPrefix = "ua:blacklist:"

	// UAPoolTTL is how long the embedded UA pool persists before refresh.
	UAPoolTTL = 7 * 24 * time.Hour

	// UARotationTTL is how often the active UA per provider rotates.
	UARotationTTL = 1 * time.Hour

	// UABlacklistTTL is the cooldown before a blacklisted UA is rehabilitated.
	UABlacklistTTL = 1 * time.Hour
)

// defaultUAPool contains modern, current User-Agent strings across major
// browsers and platforms. Must NOT contain Go runtime identifiers.
var defaultUAPool = []string{
	// Chrome (Windows)
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36",
	// Chrome (macOS)
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
	// Chrome (Linux)
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36",
	// Firefox (Windows)
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:133.0) Gecko/20100101 Firefox/133.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:132.0) Gecko/20100101 Firefox/132.0",
	// Firefox (macOS)
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:133.0) Gecko/20100101 Firefox/133.0",
	// Firefox (Linux)
	"Mozilla/5.0 (X11; Linux x86_64; rv:133.0) Gecko/20100101 Firefox/133.0",
	"Mozilla/5.0 (X11; Linux x86_64; rv:132.0) Gecko/20100101 Firefox/132.0",
	// Edge (Windows)
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36 Edg/130.0.0.0",
	// Edge (macOS)
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0",
	// Safari (macOS)
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.2 Safari/605.1.15",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1 Safari/605.1.15",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Safari/605.1.15",
	// Chrome (Windows 11)
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36",
	// Firefox ESR
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:128.0) Gecko/20100101 Firefox/128.0",
	"Mozilla/5.0 (X11; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0",
}

// uaHash returns a short hex hash of a UA string for BuntDB key use.
func uaHash(ua string) string {
	h := sha256.Sum256([]byte(ua))
	return hex.EncodeToString(h[:8])
}

// SeedUAPool writes the default UA pool to BuntDB if not already present.
func (s *Store) SeedUAPool() {
	if err := s.db.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Get(uaPoolKey)
		if err == nil {
			return nil // pool already seeded
		}
		data, marshalErr := json.Marshal(defaultUAPool)
		if marshalErr != nil {
			return marshalErr
		}
		_, _, err = tx.Set(uaPoolKey, string(data), &buntdb.SetOptions{
			Expires: true,
			TTL:     UAPoolTTL,
		})
		return err
	}); err != nil {
		slog.Warn("failed to seed UA pool", "error", err)
	}
}

// GetUA returns the active User-Agent for the given provider. If the current
// UA has expired (1h rotation) or is blacklisted, a new one is selected from
// the pool. Falls back to a random pool entry if all are blacklisted.
func (s *Store) GetUA(provider string) string {
	currentKey := uaCurrentPrefix + provider

	// Check if current UA is still valid and not blacklisted.
	var current string
	err := s.db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(currentKey)
		if err != nil {
			return err
		}
		// Verify not blacklisted.
		blKey := uaBlacklistPrefix + uaHash(val)
		if _, err := tx.Get(blKey); err == nil {
			return buntdb.ErrNotFound // blacklisted, force rotation
		}
		current = val
		return nil
	})
	if err == nil && current != "" {
		return current
	}

	// Rotate: pick a non-blacklisted UA from the pool.
	pool := s.getUAPool()
	if len(pool) == 0 {
		pool = defaultUAPool
	}

	// Shuffle to avoid predictable ordering.
	shuffled := make([]string, len(pool))
	copy(shuffled, pool)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	selected := shuffled[0] // fallback: use first even if blacklisted
	for _, ua := range shuffled {
		if !s.isBlacklisted(ua) {
			selected = ua
			break
		}
	}

	// Persist the selection.
	if err := s.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(currentKey, selected, &buntdb.SetOptions{
			Expires: true,
			TTL:     UARotationTTL,
		})
		return err
	}); err != nil {
		slog.Warn("failed to persist UA selection", "provider", provider, "error", err)
	}

	slog.Debug("UA rotated", "provider", provider, "ua", selected[:40]+"...")
	return selected
}

// BlacklistUA marks a User-Agent as temporarily blacklisted with a 1-hour
// cooldown TTL. The UA auto-rehabilitates after the TTL expires.
func (s *Store) BlacklistUA(ua string) {
	blKey := uaBlacklistPrefix + uaHash(ua)
	if err := s.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(blKey, "1", &buntdb.SetOptions{
			Expires: true,
			TTL:     UABlacklistTTL,
		})
		return err
	}); err != nil {
		slog.Warn("failed to blacklist UA", "error", err)
	}
	slog.Info("UA blacklisted (1h cooldown)", "ua_hash", uaHash(ua))
}

// getUAPool retrieves the JSON-encoded UA pool from BuntDB.
func (s *Store) getUAPool() []string {
	var pool []string
	if err := s.db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(uaPoolKey)
		if err != nil {
			return err
		}
		return json.Unmarshal([]byte(val), &pool)
	}); err != nil {
		slog.Debug("failed to read UA pool", "error", err)
	}
	return pool
}

// isBlacklisted checks if a UA is currently in the blacklist.
func (s *Store) isBlacklisted(ua string) bool {
	blKey := uaBlacklistPrefix + uaHash(ua)
	err := s.db.View(func(tx *buntdb.Tx) error {
		_, err := tx.Get(blKey)
		return err
	})
	return err == nil
}
