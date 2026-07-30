package engine

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/maccavelli/mcp-server-duckduckgo/internal/config"
	"github.com/maccavelli/mcp-server-duckduckgo/internal/models"
)

const (
	// staggerDelay is the delay between launching consecutive providers
	// to avoid simultaneous requests that trigger rate limiters.
	staggerDelay = 500 * time.Millisecond
)

// providerResult is an internal type for gathering search results from different engines.
type providerResult struct {
	data     []models.SearchResult
	err      error
	provider string
}

// SearchProvider is a standard interface for search engines.
type SearchProvider interface {
	Name() string
	Search(ctx context.Context, query string, maxResults int) ([]models.SearchResult, error)
}

func (e *SearchEngine) filterHealthyProviders(providers []SearchProvider) []SearchProvider {
	healthy := make([]SearchProvider, 0, len(providers))
	for _, p := range providers {
		if e.store.IsHealthy(p.Name()) {
			healthy = append(healthy, p)
			continue
		}
		slog.Warn("skipping unhealthy provider", "provider", p.Name())
	}

	if len(healthy) == 0 {
		slog.Warn("all providers unhealthy; attempting all as fallback")
		return providers
	}
	return healthy
}

func (e *SearchEngine) recordProviderResult(p SearchProvider, data []models.SearchResult, err error) {
	if err != nil || len(data) == 0 {
		e.store.RecordFailure(p.Name())
		return
	}
	e.store.RecordSuccess(p.Name())
}

func appendDedupedResults(
	merged []models.SearchResult,
	data []models.SearchResult,
	seen map[string]bool,
	dedupeKey func(models.SearchResult) string,
	maxResults int,
) ([]models.SearchResult, bool) {
	for _, r := range data {
		key := dedupeKey(r)
		if key != "" {
			if seen[key] {
				continue
			}
			seen[key] = true
		}
		merged = append(merged, r)
		if len(merged) >= maxResults {
			return merged, true
		}
	}
	return merged, false
}

func (e *SearchEngine) launchProviders(
	ctx context.Context,
	healthy []SearchProvider,
	query string,
	maxResults int,
	resChan chan<- providerResult,
) {
	go func() {
		for i, p := range healthy {
			if i > 0 {
				select {
				case <-ctx.Done():
					return
				case <-time.After(staggerDelay):
				}
			}

			go func(p SearchProvider) {
				data, err := p.Search(ctx, query, maxResults)
				e.recordProviderResult(p, data, err)

				select {
				case resChan <- providerResult{data, err, p.Name()}:
				case <-ctx.Done():
				}
			}(p)
		}
	}()
}

// runProviders executes multiple search providers with staggered launch,
// health-based filtering, merges results, and deduplicates them.
func (e *SearchEngine) runProviders(
	ctx context.Context,
	query string,
	maxResults int,
	dedupeKey func(models.SearchResult) string,
	providers ...SearchProvider,
) ([]models.SearchResult, error) {
	ctx, cancel := context.WithTimeout(ctx, config.DefaultTimeout)
	defer cancel()

	healthy := e.filterHealthyProviders(providers)
	resChan := make(chan providerResult, len(healthy))
	e.launchProviders(ctx, healthy, query, maxResults, resChan)

	merged := make([]models.SearchResult, 0, maxResults)
	seen := make(map[string]bool)
	received := 0

	for received < len(healthy) {
		select {
		case res := <-resChan:
			received++
			if res.err == nil {
				var done bool
				merged, done = appendDedupedResults(merged, res.data, seen, dedupeKey, maxResults)
				if done {
					return merged, nil
				}
			}
		case <-ctx.Done():
			if len(merged) > 0 {
				return merged, nil
			}
			return nil, ctx.Err()
		}
	}

	if len(merged) > 0 {
		return merged, nil
	}

	return nil, fmt.Errorf("no results found across %d providers", len(healthy))
}

type simpleProvider struct {
	name       string
	searchFunc func(context.Context, string, int) ([]models.SearchResult, error)
}

func (s *simpleProvider) Name() string { return s.name }
func (s *simpleProvider) Search(ctx context.Context, query string, maxResults int) ([]models.SearchResult, error) {
	return s.searchFunc(ctx, query, maxResults)
}
