package engine

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestNewsSearch(t *testing.T) {
	t.Run("success_ddg", func(t *testing.T) {
		jsonResp := `{"results": [{"title": "T1", "url": "U1", "excerpt": "S1", "source": "Src1", "date": "2024-01-01"}]}`
		count := 0
		client := newMockClient(func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Host, "duckduckgo.com") {
				count++
				if count == 1 {
					return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`vqd='123'`))}, nil
				}
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(jsonResp))}, nil
			}
			return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(``))}, nil
		})
		e := newTestEngine(t)
		e.Client = client
		results, err := e.NewsSearch(context.Background(), "test", 1)
		if err != nil || len(results) != 1 {
			t.Fatalf("unexpected error: %v, len=%d", err, len(results))
		}
	})

	t.Run("fallback_google_news", func(t *testing.T) {
		client := newMockClient(func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Host, "duckduckgo.com") {
				return &http.Response{StatusCode: 500, Body: io.NopCloser(strings.NewReader(`error`))}, nil
			}
			if strings.Contains(req.URL.Host, "google.com") && strings.Contains(req.URL.RawQuery, "tbm=nws") {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`<html><body><div class="g"><h3>News 1</h3><a href="http://news1.com"></a><div class="VwiC3b">Snippet 1</div></div></body></html>`)),
				}, nil
			}
			return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(``))}, nil
		})
		e := newTestEngine(t)
		e.Client = client
		results, err := e.NewsSearch(context.Background(), "test", 1)
		if err != nil || len(results) != 1 {
			t.Fatalf("fallback failed: %v, len=%d", err, len(results))
		}
		if results[0].Title != "News 1" {
			t.Errorf("expected Title News 1, got %s", results[0].Title)
		}
	})
}
