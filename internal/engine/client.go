package engine

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/maccavelli/mcp-server-duckduckgo/internal/config"
	"github.com/maccavelli/mcp-server-duckduckgo/internal/store"

	"golang.org/x/net/http2"
	"golang.org/x/sync/singleflight"
)

// Pre-compiled VQD extraction patterns.
// Primary: standard HTML attribute patterns.
// Fallback: JavaScript variable assignment patterns per Socratic correction.
var vqdPatterns = []*regexp.Regexp{
	regexp.MustCompile(`vqd='([^']+)'`),
	regexp.MustCompile(`vqd="([^"]+)"`),
	regexp.MustCompile(`vqd=([^&]+)`),
	regexp.MustCompile(`vqd_[a-zA-Z0-9]+=["']?([^"'&\s]+)`),
	regexp.MustCompile(`nrj\('/d\.js\?[^']*vqd=([^&']+)`),
}

// SearchEngine handles DDG scraping logic with BuntDB-backed persistence.
type SearchEngine struct {
	Client *http.Client
	store  *store.Store
	vqdSF  singleflight.Group
}

// NewSearchEngine initializes an optimized HTTP client for search engine
// scraping with BuntDB-backed cookie persistence and UA rotation.
func NewSearchEngine(s *store.Store) *SearchEngine {
	jar := store.NewBuntDBJar(s.DB())

	transport := &http.Transport{
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   20,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if err := http2.ConfigureTransport(transport); err != nil {
		slog.Warn("failed to configure HTTP/2 transport", "error", err)
	}

	return &SearchEngine{
		Client: &http.Client{
			Timeout:   config.DefaultTimeout,
			Transport: transport,
			Jar:       jar,
		},
		store: s,
	}
}

// Store returns the underlying store for use by handler layers (e.g. health tracking).
func (e *SearchEngine) Store() *store.Store {
	return e.store
}

// truncate safely trims a string to a maximum length of runes and adds an ellipsis.
func truncate(s string, limit int) string {
	r := []rune(s)
	if len(r) <= limit {
		return s
	}
	return string(r[:limit]) + "..."
}

// newRequest creates an HTTP request with UA rotation from the BuntDB store.
func (e *SearchEngine) newRequest(ctx context.Context, method, u string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, err
	}
	// UA rotation: pull from BuntDB store with 1h rotation per provider.
	// Provider is derived from the request hostname.
	provider := req.URL.Hostname()
	req.Header.Set("User-Agent", e.store.GetUA(provider))
	return req, nil
}

// getVQD retrieves or fetches a VQD token with BuntDB persistence as the
// single source of truth and singleflight for cache-miss deduplication.
func (e *SearchEngine) getVQD(ctx context.Context, query string) (string, error) {
	// 1. Check BuntDB first (sub-millisecond).
	if vqd, ok := e.store.GetVQD(query); ok {
		slog.Debug("VQD cache hit (BuntDB)", "query", query)
		return vqd, nil
	}

	// 2. Cache miss: use singleflight to deduplicate concurrent network fetches.
	v, err, shared := e.vqdSF.Do(query, func() (any, error) {
		// Re-check BuntDB inside singleflight to avoid races.
		if vqd, ok := e.store.GetVQD(query); ok {
			return vqd, nil
		}

		slog.Info("VQD cache miss; fetching new token", "query", query)
		var resp *http.Response
		var doErr error
		backoff := 1 * time.Second

		for i := range 3 {
			reqCtx, reqCancel := context.WithTimeout(ctx, 15*time.Second)
			req, err := e.newRequest(reqCtx, http.MethodGet, "https://duckduckgo.com", http.NoBody)
			if err != nil {
				reqCancel()
				return "", fmt.Errorf("failed to create VQD request: %w", err)
			}

			q := req.URL.Query()
			q.Add("q", query)
			req.URL.RawQuery = q.Encode()

			resp, doErr = e.Client.Do(req)
			if doErr == nil {
				if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable || resp.StatusCode == http.StatusForbidden {
					slog.Warn("VQD request rate limited or forbidden; retrying", "attempt", i+1, "status", resp.StatusCode)
					select {
					case <-ctx.Done():
						closeResponseBody(resp.Body)
						reqCancel()
						return "", ctx.Err()
					case <-time.After(backoff):
						backoff *= 2
						if i < 2 {
							closeResponseBody(resp.Body)
						}
						reqCancel()
						continue
					}
				}
				defer reqCancel()
				break
			}
			reqCancel()

			slog.Warn("VQD request failed; retrying", "attempt", i+1, "error", doErr)
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff):
				backoff *= 2
			}
		}

		if doErr != nil {
			return "", fmt.Errorf("failed to perform VQD request after retries: %w", doErr)
		}
		defer closeResponseBody(resp.Body)

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("vqd fetch failed with status code: %d", resp.StatusCode)
		}

		body, err := io.ReadAll(io.LimitReader(resp.Body, config.MaxBodyBytes))
		if err != nil {
			return "", fmt.Errorf("failed to read VQD response body: %w", err)
		}

		for _, re := range vqdPatterns {
			matches := re.FindSubmatch(body)
			if len(matches) > 1 {
				vqd := string(matches[1])
				// Persist to BuntDB.
				e.store.SetVQD(query, vqd, store.VQDDefaultTTL)
				return vqd, nil
			}
		}

		return "", fmt.Errorf("could not extract vqd")
	})

	if err != nil {
		return "", err
	}

	if shared {
		slog.Debug("VQD fetch shared across concurrent requests", "query", query)
	}

	vqd, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("unexpected VQD cache value type %T", v)
	}
	return vqd, nil
}
