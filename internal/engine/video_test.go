package engine

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestVideoSearch(t *testing.T) {
	t.Run("success_ddg", func(t *testing.T) {
		jsonResp := `{"results": [{"title": "V1", "description": "D1"}]}`
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
		results, err := e.VideoSearch(context.Background(), "test", 1)
		if err != nil || len(results) != 1 {
			t.Fatalf("unexpected error: %v, len=%d", err, len(results))
		}
	})

	t.Run("fallback_google_video", func(t *testing.T) {
		client := newMockClient(func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Host, "duckduckgo.com") {
				return &http.Response{StatusCode: 500, Body: io.NopCloser(strings.NewReader(`error`))}, nil
			}
			if strings.Contains(req.URL.Host, "google.com") && strings.Contains(req.URL.RawQuery, "tbm=vid") {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`<html><body><div class="g"><h3>Video 1</h3><a href="http://video1.com"></a><div class="VwiC3b">Snippet 1</div></div></body></html>`)),
				}, nil
			}
			return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(``))}, nil
		})
		e := newTestEngine(t)
		e.Client = client
		results, err := e.VideoSearch(context.Background(), "test", 1)
		if err != nil || len(results) != 1 {
			t.Fatalf("fallback failed: %v, len=%d", err, len(results))
		}
		if results[0].Title != "Video 1" {
			t.Errorf("expected Title Video 1, got %s", results[0].Title)
		}
	})
}
