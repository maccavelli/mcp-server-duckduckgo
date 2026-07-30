package engine

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestWebSearch(t *testing.T) {
	t.Run("concurrent_ddg_and_google", func(t *testing.T) {
		ddgHtml := `<div class="web-result"><a class="result__a" href="http://ddg.com">DDG Result</a><div class="result__title">DDG Result</div><div class="result__snippet">DDG Snippet</div></div>`
		googleHtml := `<html><body><div class="g"><h3>Google Result</h3><a href="http://google.com"></a><div class="VwiC3b">Google Snippet</div></div></body></html>`

		client := newMockClient(func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Host, "duckduckgo.com") {
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(ddgHtml))}, nil
			}
			if strings.Contains(req.URL.Host, "google.com") {
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(googleHtml))}, nil
			}
			return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(``))}, nil
		})

		e := newTestEngine(t)
		e.Client = client
		results, err := e.WebSearch(context.Background(), "test", 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Expecting at least 2 results (one from each provider)
		if len(results) < 2 {
			t.Fatalf("expected at least 2 results from concurrent search, got %d", len(results))
		}
	})

	t.Run("deduplication", func(t *testing.T) {
		commonUrl := "http://common.com"
		ddgHtml := `<div class="web-result"><a class="result__a" href="` + commonUrl + `">Title</a><div class="result__title">Title</div></div>`
		googleHtml := `<html><body><div class="g"><h3>Title</h3><a href="` + commonUrl + `"></a></div></body></html>`

		client := newMockClient(func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Host, "duckduckgo.com") {
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(ddgHtml))}, nil
			}
			if strings.Contains(req.URL.Host, "google.com") {
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(googleHtml))}, nil
			}
			return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(``))}, nil
		})

		e := newTestEngine(t)
		e.Client = client
		results, err := e.WebSearch(context.Background(), "test", 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Expecting only 1 result due to deduplication
		if len(results) != 1 {
			t.Errorf("expected 1 result after deduplication, got %d", len(results))
		}
	})

	t.Run("fallback_bing", func(t *testing.T) {
		bingHtml := `<html><body><li class="b_algo"><h2>Bing Result</h2><a href="http://bing.com"></a><div class="b_caption"><p>Bing Snippet</p></div></li></body></html>`
		client := newMockClient(func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Host, "duckduckgo.com") || strings.Contains(req.URL.Host, "google.com") {
				return &http.Response{StatusCode: 500, Body: io.NopCloser(strings.NewReader(`error`))}, nil
			}
			if strings.Contains(req.URL.Host, "bing.com") {
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(bingHtml))}, nil
			}
			return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(``))}, nil
		})

		e := newTestEngine(t)
		e.Client = client
		results, err := e.WebSearch(context.Background(), "test", 5)
		if err != nil || len(results) == 0 {
			t.Fatalf("fallback to bing failed: %v, len=%d", err, len(results))
		}
		if results[0].Source != "Bing" {
			t.Errorf("expected source Bing, got %s", results[0].Source)
		}
	})
}
