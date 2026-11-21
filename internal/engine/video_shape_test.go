package engine

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/maccavelli/mcp-server-duckduckgo/internal/models"
)

func runVideoShapeTests(t *testing.T) {
	t.Run("content_url", func(t *testing.T) {
		body := `{"results":[{"title":"How to fold an Origami Crane","content":"https://www.youtube.com/watch?v=abc","embed_url":"https://www.youtube.com/embed/abc","duration":"3:56","publisher":"YouTube"}]}`
		results := videoOnce(t, body)
		if len(results) != 1 {
			t.Fatalf("len %d", len(results))
		}
		if results[0].URL != "https://www.youtube.com/watch?v=abc" {
			t.Fatalf("url %q", results[0].URL)
		}
		if results[0].URL == "https://www.youtube.com/embed/abc" {
			t.Fatal("url is the embed URL")
		}
		if results[0].Duration != "3:56" || results[0].Publisher != "YouTube" {
			t.Fatalf("duration %q publisher %q", results[0].Duration, results[0].Publisher)
		}
	})

	t.Run("embed_url_fallback", func(t *testing.T) {
		body := `{"results":[{"title":"Embed","content":"","embed_url":"https://www.youtube.com/embed/abc"}]}`
		results := videoOnce(t, body)
		if len(results) != 1 || results[0].URL != "https://www.youtube.com/embed/abc" {
			t.Fatalf("results %+v", results)
		}
	})

	t.Run("no_rows_skips_bing", func(t *testing.T) {
		var mu sync.Mutex
		var hosts []string
		client := newMockClient(func(req *http.Request) (*http.Response, error) {
			mu.Lock()
			hosts = append(hosts, req.URL.Host)
			mu.Unlock()
			if strings.Contains(req.URL.Path, "v.js") {
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"results":[]}`))}, nil
			}
			if strings.Contains(req.URL.Host, "duckduckgo.com") {
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`vqd='123'`))}, nil
			}
			if strings.Contains(req.URL.RawQuery, "tbm=vid") {
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`<html><body><p>no cards</p></body></html>`))}, nil
			}
			return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(""))}, nil
		})
		e := newTestEngine(t)
		e.Client = client
		_, err := e.VideoSearch(context.Background(), "origami", 1)
		if err == nil {
			t.Fatal("expected an error when both video providers are empty")
		}
		mu.Lock()
		defer mu.Unlock()
		for _, h := range hosts {
			if strings.Contains(h, "bing.com") {
				t.Fatalf("bing host %q in %v", h, hosts)
			}
		}
	})
}

func videoOnce(t *testing.T, body string) []models.SearchResult {
	t.Helper()
	count := 0
	client := newMockClient(func(req *http.Request) (*http.Response, error) {
		if strings.Contains(req.URL.Host, "duckduckgo.com") {
			count++
			if strings.Contains(req.URL.Path, "v.js") || count > 1 {
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
			}
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`vqd='123'`))}, nil
		}
		return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(""))}, nil
	})
	e := newTestEngine(t)
	e.Client = client
	results, err := e.VideoSearch(context.Background(), "origami", 1)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	return results
}
