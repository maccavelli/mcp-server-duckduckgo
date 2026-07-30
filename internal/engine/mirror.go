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

// BookSearch performs a book search using Anna's Archive for functional parity.
// It queries multiple mirrors concurrently for improved performance and resilience.
func (e *SearchEngine) BookSearch(ctx context.Context, query string, maxResults int) ([]models.SearchResult, error) {
	mirrors := []string{"gd", "gl", "pk"}

	type result struct {
		data []models.SearchResult
		err  error
	}

	resChan := make(chan result, len(mirrors))
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, tld := range mirrors {
		go func(tld string) {
			data, err := e.queryMirror(childCtx, tld, query, maxResults)
			select {
			case resChan <- result{data, err}:
				if err == nil && len(data) > 0 {
					cancel() // Stop other requests if we found results
				}
			case <-childCtx.Done():
				// Already finished or cancelled
			}
		}(tld)
	}

	var lastErr error
	for range mirrors {
		select {
		case res := <-resChan:
			if res.err == nil && len(res.data) > 0 {
				return res.data, nil
			}
			if res.err != nil {
				lastErr = res.err
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("book search failed across all mirrors: %w", lastErr)
	}
	return nil, fmt.Errorf("book search returned no results from any mirror")
}

// queryMirror performs a search against a specific Anna's Archive mirror.
func (e *SearchEngine) queryMirror(ctx context.Context, tld, query string, maxResults int) ([]models.SearchResult, error) {
	baseUrl := fmt.Sprintf("https://annas-archive.%s", tld)
	u := fmt.Sprintf("%s/search?q=%s", baseUrl, url.QueryEscape(query))

	respBody, err := e.fetchMirrorHTML(ctx, u)
	if err != nil {
		return nil, err
	}

	return e.parseMirrorHTML(baseUrl, respBody, maxResults)
}

func (e *SearchEngine) fetchMirrorHTML(ctx context.Context, u string) ([]byte, error) {
	req, err := e.newRequest(ctx, "GET", u, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := e.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer closeResponseBody(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mirror returned status %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, config.MaxBodyBytes))
	if err != nil {
		return nil, err
	}

	// Anna's Archive hides results in comments, need to strip them
	respBody = bytes.ReplaceAll(respBody, []byte("<!--"), nil)
	respBody = bytes.ReplaceAll(respBody, []byte("-->"), nil)
	return respBody, nil
}

func (e *SearchEngine) parseMirrorHTML(baseUrl string, htmlData []byte, maxResults int) ([]models.SearchResult, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(htmlData)))
	if err != nil {
		return nil, err
	}

	results := make([]models.SearchResult, 0, maxResults)
	doc.Find("div.js-aarecord-list-outer div.flex").Each(func(i int, s *goquery.Selection) {
		if len(results) >= maxResults {
			return
		}

		res := e.extractMirrorResult(baseUrl, s)
		if res.Title != "" && res.URL != "" {
			// Check for duplicates
			isDuplicate := false
			for _, r := range results {
				if r.URL == res.URL {
					isDuplicate = true
					break
				}
			}
			if !isDuplicate {
				results = append(results, res)
			}
		}
	})

	return results, nil
}

func (e *SearchEngine) extractMirrorResult(baseUrl string, s *goquery.Selection) models.SearchResult {
	anchor := s.Find("a.text-lg, a.font-semibold, a.js-vim-focus").First()
	if anchor.Length() == 0 {
		anchor = s.Find("a").First()
	}
	title := strings.TrimSpace(anchor.Text())

	link, exists := anchor.Attr("href")
	if !exists || link == "" {
		link, exists = s.Find("a").First().Attr("href")
		if !exists {
			return models.SearchResult{}
		}
	}

	author := ""
	s.Find("a").Each(func(i int, sel *goquery.Selection) {
		if sel.Find(".icon-\\[mdi--user-edit\\]").Length() > 0 {
			author = strings.TrimSpace(sel.Text())
		}
	})
	if author == "" {
		author = strings.TrimSpace(s.Find("div.italic").Text())
	}

	var infoLine string
	var description string
	s.Find("div.text-gray-800, div.text-sm.text-gray-600").Each(func(i int, sel *goquery.Selection) {
		sel.Find("script, style").Remove()
		txt := strings.TrimSpace(sel.Text())
		if txt == "" {
			return
		}
		if strings.Contains(txt, "·") && infoLine == "" {
			infoLine = txt
		} else if description == "" || len(txt) > len(description) {
			description = txt
		}
	})

	href := link
	if href != "" && !strings.HasPrefix(href, "http") {
		href = baseUrl + href
	}

	return models.SearchResult{
		Title:       title,
		URL:         href,
		Description: truncate(description, config.MaxSnippetLength),
		Author:      truncate(author, config.MaxSnippetLength),
		Info:        truncate(infoLine, config.MaxSnippetLength),
	}
}
