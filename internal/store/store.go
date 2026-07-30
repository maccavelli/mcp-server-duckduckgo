// Package store provides BuntDB-backed persistent storage for the DDG search
// scraping resilience engine. It manages VQD token caching, User-Agent rotation,
// cookie persistence, and provider health tracking with automatic TTL eviction.
package store

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/tidwall/buntdb"

	mcpbuntdb "github.com/maccavelli/mcp-buntdb"
)

const (
	// dbFileName is the BuntDB database file within the data directory.
	dbFileName = "scraping.db"

	// shrinkInterval is how often the background goroutine compacts the
	// append-only log to physically remove expired entries.
	shrinkInterval = 6 * time.Hour
)

// Store wraps a BuntDB instance providing TTL-managed persistence for
// scraping state (VQD tokens, cookies, UA rotation, provider health).
type Store struct {
	db     *buntdb.DB
	cancel context.CancelFunc
}

// Open initialises the BuntDB store at the given directory path. If the
// directory does not exist it is created with 0700 permissions. The database
// file is created with 0600 permissions. On any failure the store degrades
// gracefully to an in-memory BuntDB instance.
func Open(dataDir string) (*Store, error) {
	dbPath := filepath.Join(dataDir, dbFileName)

	cfg := buntdb.Config{
		AutoShrinkPercentage: 50,
		AutoShrinkMinSize:    25 * 1024 * 1024,
	}

	db, err := mcpbuntdb.OpenBuntDB(dbPath, &cfg)
	if err != nil {
		slog.Warn("failed to open BuntDB file; falling back to in-memory mode",
			"path", dbPath, "error", err)
		return openInMemory()
	}

	// Restrict the file to owner-only after BuntDB creates it.
	if dbPath != ":memory:" {
		if err := os.Chmod(dbPath, 0600); err != nil {
			slog.Warn("failed to chmod BuntDB file", "path", dbPath, "error", err)
		}
	}

	// STD-MCP-CONTEXT-DETACHMENT-001: Background shrink goroutine uses a
	// detached context so it is NOT tied to any request lifecycle.
	ctx, cancel := context.WithCancel(context.Background())
	s := &Store{db: db, cancel: cancel}
	go s.shrinkLoop(ctx)
	return s, nil
}

// openInMemory creates a volatile in-memory BuntDB instance that preserves
// all API contracts but does not survive process restarts.
func openInMemory() (*Store, error) {
	db, err := buntdb.Open(":memory:")
	if err != nil {
		return nil, err
	}
	slog.Warn("operating in in-memory mode — scraping state will not persist across restarts")
	ctx, cancel := context.WithCancel(context.Background())
	s := &Store{db: db, cancel: cancel}
	go s.shrinkLoop(ctx)
	return s, nil
}

// Close shuts down the periodic shrink goroutine and closes the database.
// STD-MCP-CLI-LIFECYCLE-001: callers MUST defer Close() immediately after Open().
func (s *Store) Close() error {
	s.cancel()
	return s.db.Close()
}

// DB exposes the underlying BuntDB instance for direct use by sub-packages
// (e.g. the CookieJar implementation).
func (s *Store) DB() *buntdb.DB {
	return s.db
}

// shrinkLoop periodically compacts the append-only log to physically remove
// expired entries. This is important for security: expired cookies remain in
// the physical file until compaction.
func (s *Store) shrinkLoop(ctx context.Context) {
	ticker := time.NewTicker(shrinkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.db.Shrink(); err != nil {
				slog.Warn("BuntDB shrink failed", "error", err)
			} else {
				slog.Debug("BuntDB shrink completed successfully")
			}
		}
	}
}
