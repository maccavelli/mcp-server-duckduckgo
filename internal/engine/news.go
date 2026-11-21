package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/maccavelli/mcp-server-duckduckgo/internal/config"
	"github.com/maccavelli/mcp-server-duckduckgo/internal/models"
)

var newsTagRe = regexp.MustCompile(`(?i)<[^>]+>`)

var newsPhrases = []struct {
	text string
	df   string
}{
	{"right now", "d"},
	{"today", "d"},
	{"this week", "w"},
	{"this month", "m"},
}

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

	if _, _, err := prepareNewsQuery(query); err != nil {
		return nil, err
	}

	return e.runProviders(ctx, query, maxResults, dedupeKey, providers...)
}

func (e *SearchEngine) ddgNewsSearch(ctx context.Context, query string, maxResults int) ([]models.SearchResult, error) {
	q, df, err := prepareNewsQuery(query)
	if err != nil {
		return nil, err
	}

	vqd, err := e.getVQD(ctx, q)
	if err != nil {
		return nil, err
	}

	vals := url.Values{}
	vals.Set("q", q)
	vals.Set("vqd", vqd)
	vals.Set("l", "us-en")
	vals.Set("o", "json")
	vals.Set("p", "-1")
	vals.Set("noamp", "1")
	if df != "" {
		vals.Set("df", df)
	}
	u := "https://duckduckgo.com/news.js?" + vals.Encode()
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
			Description: plainNewsExcerpt(r.Excerpt),
			Source:      r.Source,
			Date:        dateStr,
			ImageURL:    r.Image,
		})
	}
	return results, nil
}

func prepareNewsQuery(raw string) (string, string, error) {
	runes := []rune(strings.Join(strings.Fields(raw), " "))
	type span struct{ start, end int }
	var spans []span
	var dfs []string
	for i := 0; i < len(runes); {
		if i > 0 && newsWordRune(runes[i-1]) {
			i++
			continue
		}
		matched := false
		for _, phrase := range newsPhrases {
			pr := []rune(phrase.text)
			end := i + len(pr)
			if end > len(runes) || !newsEqualFold(runes[i:end], pr) {
				continue
			}
			if phrase.text == "today" && end+1 < len(runes) && (runes[end] == '\'' || runes[end] == '\u2019') && runes[end+1] == 's' {
				end += 2
			}
			if end < len(runes) && newsWordRune(runes[end]) {
				continue
			}
			spans = append(spans, span{start: i, end: end})
			dfs = append(dfs, phrase.df)
			i = end
			matched = true
			break
		}
		if !matched {
			i++
		}
	}

	df := ""
	for _, candidate := range []string{"d", "w", "m"} {
		for _, got := range dfs {
			if got == candidate {
				df = candidate
				break
			}
		}
		if df != "" {
			break
		}
	}

	kept := make([]rune, 0, len(runes))
	at := 0
	for _, s := range spans {
		kept = append(kept, runes[at:s.start]...)
		at = s.end
	}
	kept = append(kept, runes[at:]...)
	q := strings.Join(strings.Fields(string(kept)), " ")
	if q == "" {
		return "", "", fmt.Errorf("news query is empty")
	}
	return q, df, nil
}

func newsWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func newsEqualFold(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if unicode.ToLower(a[i]) != unicode.ToLower(b[i]) {
			return false
		}
	}
	return true
}

func plainNewsExcerpt(s string) string {
	s = html.UnescapeString(s)
	s = newsTagRe.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	s = strings.Join(strings.Fields(s), " ")
	return truncate(s, config.MaxSnippetLength)
}
