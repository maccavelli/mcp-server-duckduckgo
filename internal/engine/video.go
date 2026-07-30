package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/maccavelli/mcp-server-duckduckgo/internal/config"
	"github.com/maccavelli/mcp-server-duckduckgo/internal/models"
)

// VideoSearch performs a high-concurrency video search across multiple providers.
func (e *SearchEngine) VideoSearch(ctx context.Context, query string, maxResults int) ([]models.SearchResult, error) {
	providers := []SearchProvider{
		&simpleProvider{name: "DuckDuckGo Videos", searchFunc: e.ddgVideoSearch},
		&simpleProvider{name: "Google Videos", searchFunc: func(c context.Context, q string, m int) ([]models.SearchResult, error) {
			return e.GoogleSearch(c, q, "vid", m)
		}},
	}

	dedupeKey := func(r models.SearchResult) string {
		return r.URL
	}

	results, err := e.runProviders(ctx, query, maxResults, dedupeKey, providers...)
	if err == nil {
		return results, nil
	}

	return e.BingWebSearch(ctx, query, maxResults)
}

func (e *SearchEngine) ddgVideoSearch(ctx context.Context, query string, maxResults int) ([]models.SearchResult, error) {
	vqd, err := e.getVQD(ctx, query)
	if err != nil {
		return nil, err
	}

	u := fmt.Sprintf("https://duckduckgo.com/v.js?q=%s&vqd=%s&o=json&l=us-en&p=-1", url.QueryEscape(query), vqd)
	req, err := e.newRequest(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", "https://duckduckgo.com/")

	resp, err := e.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer closeResponseBody(resp.Body)

	if isJSONEndpointServingHTML(resp) {
		e.store.RecordCAPTCHA("DuckDuckGo Videos")
		return nil, ErrCAPTCHADetected
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ddg video search failed with status %d", resp.StatusCode)
	}

	var data struct {
		Results []struct {
			Title       string `json:"title"`
			URL         string `json:"url"`
			Description string `json:"description"`
			Duration    string `json:"duration"`
			Publisher   string `json:"publisher"`
			Published   string `json:"published"`
		} `json:"results"`
	}

	if err := json.NewDecoder(io.LimitReader(resp.Body, config.MaxBodyBytes)).Decode(&data); err != nil {
		return nil, err
	}

	results := make([]models.SearchResult, 0, maxResults)
	for i, r := range data.Results {
		if i >= maxResults {
			break
		}
		results = append(results, models.SearchResult{
			Title:       r.Title,
			URL:         r.URL,
			Description: truncate(r.Description, config.MaxSnippetLength),
			Duration:    r.Duration,
			Publisher:   r.Publisher,
			Date:        r.Published,
		})
	}
	return results, nil
}
