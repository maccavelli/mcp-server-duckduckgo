package engine

import (
	"context"
	"io"

	"net/http"

	"strings"
	"testing"
)

func TestBookSearch_Coverage(t *testing.T) {
	// Structural HTML that matches mirror.go's scraping logic
	mockHTML := `
		<div class="js-aarecord-list-outer">
			<div class="flex">
				<a class="js-vim-focus" href="/md5/123">Test Book Title</a>
				<div class="italic">Test Author</div>
				<div class="text-gray-800">2024 · English · 1MB · pdf</div>
			</div>
		</div>`

	t.Run("both_mirrors_success", func(t *testing.T) {
		client := newMockClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(mockHTML)),
			}, nil
		})
		e := newTestEngine(t)
		e.Client = client
		results, err := e.BookSearch(context.Background(), "test", 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("expected results from mirrors")
		}
		if results[0].Title != "Test Book Title" {
			t.Errorf("expected 'Test Book Title', got %s", results[0].Title)
		}
	})

	t.Run("one_mirror_failure", func(t *testing.T) {
		client := newMockClient(func(req *http.Request) (*http.Response, error) {
			if strings.Contains(req.URL.Host, "annas-archive.gd") {
				return nil, http.ErrHandlerTimeout
			}
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(mockHTML)),
			}, nil
		})
		e := newTestEngine(t)
		e.Client = client
		results, err := e.BookSearch(context.Background(), "test", 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) == 0 {
			t.Error("expected results from the working mirror")
		}
	})

	t.Run("all_mirrors_failure", func(t *testing.T) {
		client := newMockClient(func(req *http.Request) (*http.Response, error) {
			return nil, http.ErrHandlerTimeout
		})
		e := newTestEngine(t)
		e.Client = client
		_, err := e.BookSearch(context.Background(), "test", 10)
		if err == nil || !strings.Contains(err.Error(), "book search failed across all mirrors") {
			t.Errorf("expected total failure error, got %v", err)
		}
	})

	t.Run("query_mirror_status_error", func(t *testing.T) {
		client := newMockClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader(""))}, nil
		})
		e := newTestEngine(t)
		e.Client = client
		_, err := e.BookSearch(context.Background(), "test", 10)
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("query_mirror_read_error", func(t *testing.T) {
		client := newMockClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: io.NopCloser(&errorReader{})}, nil
		})
		e := newTestEngine(t)
		e.Client = client
		_, err := e.BookSearch(context.Background(), "test", 10)
		if err == nil {
			t.Error("expected error")
		}
	})
}
