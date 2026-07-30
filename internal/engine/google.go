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

// GoogleSearch performs a search using Google's public endpoint as a fallback.
func (e *SearchEngine) GoogleSearch(ctx context.Context, query string, tbm string, maxResults int) ([]models.SearchResult, error) {
	u := fmt.Sprintf("https://www.google.com/search?q=%s&tbm=%s", url.QueryEscape(query), tbm)
	req, err := e.newRequest(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create google search request: %w", err)
	}

	resp, err := e.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform google search: %w", err)
	}
	defer closeResponseBody(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google search failed with status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, config.MaxBodyBytes))
	if err != nil {
		return nil, err
	}
	if challenge := detectGoogleChallenge(bodyBytes); challenge != "" {
		e.store.RecordCAPTCHA("Google")
		return nil, ErrCAPTCHADetected
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to parse google search results: %w", err)
	}

	if tbm == "isch" {
		return e.parseGoogleImageResults(doc, maxResults), nil
	}

	return e.parseGoogleWebResults(doc, maxResults), nil
}

func (e *SearchEngine) parseGoogleWebResults(doc *goquery.Document, maxResults int) []models.SearchResult {
	results := make([]models.SearchResult, 0, maxResults)
	doc.Find("div.g").Each(func(i int, s *goquery.Selection) {
		if len(results) >= maxResults {
			return
		}

		title := strings.TrimSpace(s.Find("h3").Text())
		link, exists := s.Find("a").Attr("href")
		snippet := strings.TrimSpace(s.Find(".VwiC3b, .y67o5b, .MUF9Of").Text())

		if title != "" && exists && link != "" && strings.HasPrefix(link, "http") {
			results = append(results, models.SearchResult{
				Title:       title,
				URL:         link,
				Description: truncate(snippet, config.MaxSnippetLength),
			})
		}
	})
	return results
}

func (e *SearchEngine) parseGoogleImageResults(doc *goquery.Document, maxResults int) []models.SearchResult {
	results := make([]models.SearchResult, 0, maxResults)
	// Google's older image search layout (still used by many scrapers)
	doc.Find("img[data-src], img[src]").Each(func(i int, s *goquery.Selection) {
		if len(results) >= maxResults {
			return
		}

		imgSrc, exists := s.Attr("data-src")
		if !exists || imgSrc == "" {
			imgSrc, exists = s.Attr("src")
			if !exists {
				return
			}
		}

		// Avoid common tracking pixels
		if imgSrc == "" || strings.HasPrefix(imgSrc, "data:image/gif") || strings.Contains(imgSrc, "/favicon.ico") {
			return
		}

		results = append(results, models.SearchResult{
			Title:    "Image result",
			ImageURL: imgSrc,
			Source:   providerGoogleImages,
		})
	})
	return results
}
