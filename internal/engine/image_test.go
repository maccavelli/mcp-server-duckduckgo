package engine

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestImageSearch(t *testing.T) {
	t.Run("success_google", func(t *testing.T) {
		client := newMockClient(func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Host, "google.com") {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`<html><body><img src="http://img1.jpg"></body></html>`)),
				}, nil
			}
			return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(``))}, nil
		})
		e := newTestEngine(t)
		e.Client = client
		results, err := e.ImageSearch(context.Background(), "test", 1)
		if err != nil || len(results) != 1 {
			t.Errorf("image search failed: %v, len=%d", err, len(results))
		}
		if results[0].Source != "Google Images" {
			t.Errorf("expected source Google Images, got %s", results[0].Source)
		}
	})

	t.Run("fallback_bing", func(t *testing.T) {
		client := newMockClient(func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Host, "google.com") {
				return &http.Response{StatusCode: 500, Body: io.NopCloser(strings.NewReader(`error`))}, nil
			}
			if strings.Contains(req.URL.Host, "bing.com") {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`<html><body><img class="mimg" src="http://img2.jpg"></body></html>`)),
				}, nil
			}
			return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(``))}, nil
		})
		e := newTestEngine(t)
		e.Client = client
		results, err := e.ImageSearch(context.Background(), "test", 1)
		if err != nil || len(results) != 1 {
			t.Errorf("image search fallback failed: %v, len=%d", err, len(results))
		}
		if results[0].Source != "Bing Images" {
			t.Errorf("expected source Bing Images, got %s", results[0].Source)
		}
	})

	t.Run("fallback_ddg", func(t *testing.T) {
		count := 0
		client := newMockClient(func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Host, "google.com") || strings.Contains(req.URL.Host, "bing.com") {
				return &http.Response{StatusCode: 500, Body: io.NopCloser(strings.NewReader(`error`))}, nil
			}
			if strings.Contains(req.URL.Host, "duckduckgo.com") {
				count++
				if count == 1 {
					return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`vqd='123'`))}, nil
				}
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`{"results": [{"title": "I1", "image": "Img1", "source": "DDG"}]}`)),
				}, nil
			}
			return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(``))}, nil
		})
		e := newTestEngine(t)
		e.Client = client
		results, err := e.ImageSearch(context.Background(), "test", 1)
		if err != nil || len(results) != 1 {
			t.Errorf("image search final fallback failed: %v, len=%d", err, len(results))
		}
		if results[0].Source != "DDG" {
			t.Errorf("expected source DDG, got %s", results[0].Source)
		}
	})
}
