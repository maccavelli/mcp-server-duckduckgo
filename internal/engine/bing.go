package engine

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/maccavelli/mcp-server-duckduckgo/internal/config"
	"github.com/maccavelli/mcp-server-duckduckgo/internal/models"

	"github.com/PuerkitoBio/goquery"
)

// BingImageSearch performs an image search using Bing's public endpoint as a fallback.
func (e *SearchEngine) BingImageSearch(ctx context.Context, query string, maxResults int) ([]models.SearchResult, error) {
	u := fmt.Sprintf("https://www.bing.com/images/search?q=%s", url.QueryEscape(query))
	req, err := e.newRequest(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create bing image search request: %w", err)
	}

	resp, err := e.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform bing image search: %w", err)
	}
	defer closeResponseBody(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bing image search failed with status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, config.MaxBodyBytes))
	if err != nil {
		return nil, err
	}
	if challenge := detectBingChallenge(bodyBytes); challenge != "" {
		e.store.RecordCAPTCHA(providerBingImages)
		return nil, ErrCAPTCHADetected
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to parse bing image search results: %w", err)
	}

	results := make([]models.SearchResult, 0, maxResults)
	doc.Find("img.mimg, div.imgpt img").Each(func(i int, s *goquery.Selection) {
		if len(results) >= maxResults {
			return
		}

		imgSrc, exists := s.Attr("src")
		if !exists || imgSrc == "" {
			imgSrc, exists = s.Attr("data-src")
			if !exists {
				return
			}
		}

		if imgSrc != "" && strings.HasPrefix(imgSrc, "http") {
			results = append(results, models.SearchResult{
				Title:    "Image result",
				ImageURL: imgSrc,
				Source:   providerBingImages,
			})
		}
	})

	return results, nil
}
