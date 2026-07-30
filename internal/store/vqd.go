package store

import (
	"encoding/base64"
	"log/slog"
	"time"

	"github.com/tidwall/buntdb"
)

const (
	// vqdPrefix is the BuntDB key prefix for VQD tokens.
	vqdPrefix = "vqd:"

	// VQDDefaultTTL is the default time-to-live for cached VQD tokens.
	VQDDefaultTTL = 5 * time.Minute
)

// vqdKey builds a BuntDB key from a raw query string using URL-safe
// base64 encoding (reversible, shorter than SHA-256, debuggable).
func vqdKey(query string) string {
	return vqdPrefix + base64.URLEncoding.EncodeToString([]byte(query))
}

// GetVQD retrieves a cached VQD token for the given query. Returns the
// token and true if found, or an empty string and false if the key is
// missing or expired (BuntDB handles TTL eviction automatically).
func (s *Store) GetVQD(query string) (string, bool) {
	var vqd string
	err := s.db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(vqdKey(query))
		if err != nil {
			return err
		}
		vqd = val
		return nil
	})
	if err != nil {
		return "", false
	}
	return vqd, true
}

// SetVQD persists a VQD token for the given query with the specified TTL.
func (s *Store) SetVQD(query, vqd string, ttl time.Duration) {
	if err := s.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(vqdKey(query), vqd, &buntdb.SetOptions{
			Expires: true,
			TTL:     ttl,
		})
		return err
	}); err != nil {
		slog.Warn("failed to persist VQD token", "error", err)
	}
}
