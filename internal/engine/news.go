package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/maccavelli/mcp-server-duckduckgo/internal/config"
	"github.com/maccavelli/mcp-server-duckduckgo/internal/models"
)

// NewsSearch performs a high-concurrency news search across multiple providers.
func (e *SearchEngine) NewsSearch(ctx context.Context, query string, maxResults int) ([]models.SearchResult, error) {
	providers := []SearchProvider{
		&simpleProvider{name: "DuckDuckGo News", searchFunc: e.ddgNewsSearch},
		&simpleProvider{name: "Google News", searchFunc: func(c context.Context, q string, m int) ([]models.SearchResult, error) {
			return e.GoogleSearch(c, q, "nws", m)
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

func (e *SearchEngine) ddgNewsSearch(ctx context.Context, query string, maxResults int) ([]models.SearchResult, error) {
	vqd, err := e.getVQD(ctx, query)
	if err != nil {
		return nil, err
	}

	u := fmt.Sprintf("https://duckduckgo.com/news.js?q=%s&vqd=%s&l=us-en&o=json&p=-1", url.QueryEscape(query), vqd)
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
		e.store.RecordCAPTCHA("DuckDuckGo News")
		return nil, ErrCAPTCHADetected
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ddg news search failed with status %d", resp.StatusCode)
	}

	var data struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Excerpt string `json:"excerpt"`
			Source  string `json:"source"`
			Date    any    `json:"date"`
			Image   string `json:"image"`
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

		dateStr := ""
		switch v := r.Date.(type) {
		case string:
			dateStr = v
		case float64:
			dateStr = time.Unix(int64(v), 0).UTC().Format(time.RFC3339)
		case int64:
			dateStr = time.Unix(v, 0).UTC().Format(time.RFC3339)
		}

		results = append(results, models.SearchResult{
			Title:       r.Title,
			URL:         r.URL,
			Description: truncate(r.Excerpt, config.MaxSnippetLength),
			Source:      r.Source,
			Date:        dateStr,
			ImageURL:    r.Image,
		})
	}
	return results, nil
}
