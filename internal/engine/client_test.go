package engine

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/maccavelli/mcp-server-duckduckgo/internal/store"
)

// newTestEngine creates a SearchEngine backed by an in-memory BuntDB store.
func newTestEngine(t *testing.T) *SearchEngine {
	t.Helper()
	s, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatalf("failed to open test store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	s.SeedUAPool()
	return NewSearchEngine(s)
}

func TestNewSearchEngine(t *testing.T) {
	e := newTestEngine(t)
	if e == nil {
		t.Fatal("expected engine")
	}
	if e.Client == nil {
		t.Fatal("expected client")
	}
}

func TestNewRequest(t *testing.T) {
	e := newTestEngine(t)
	req, err := e.newRequest(context.Background(), "GET", "https://example.com", http.NoBody)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ua := req.Header.Get("User-Agent")
	if ua == "" {
		t.Error("expected non-empty User-Agent from UA rotation")
	}
	if strings.Contains(ua, "Go-http-client") {
		t.Error("UA must not contain Go runtime identifier")
	}
}

func TestGetVQD(t *testing.T) {
	t.Run("cache_hit", func(t *testing.T) {
		count := 0
		client := newMockClient(func(req *http.Request) (*http.Response, error) {
			count++
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`vqd='123456789'`)),
			}, nil
		})
		e := newTestEngine(t)
		e.Client = client

		// First call - should hit network
		vqd1, err := e.getVQD(context.Background(), "test")
		if err != nil || vqd1 != "123456789" {
			t.Fatalf("first call failed: %v", err)
		}

		// Second call - should hit BuntDB cache
		vqd2, err := e.getVQD(context.Background(), "test")
		if err != nil || vqd2 != "123456789" {
			t.Fatalf("second call failed: %v", err)
		}

		if count != 1 {
			t.Errorf("expected 1 network call, got %d", count)
		}
	})

	t.Run("failure_not_found", func(t *testing.T) {
		client := newMockClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`no vqd here`)),
			}, nil
		})
		e := newTestEngine(t)
		e.Client = client
		_, err := e.getVQD(context.Background(), "test")
		if err == nil || !strings.Contains(err.Error(), "could not extract vqd") {
			t.Errorf("expected extraction error, got %v", err)
		}
	})

	t.Run("network_error", func(t *testing.T) {
		client := newMockClient(func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("network down")
		})
		e := newTestEngine(t)
		e.Client = client
		_, err := e.getVQD(context.Background(), "test")
		if err == nil || !strings.Contains(err.Error(), "network down") {
			t.Errorf("expected network error, got %v", err)
		}
	})
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		limit    int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hello..."},
		{"咖啡", 1, "咖..."},
		{"咖啡", 2, "咖啡"},
	}

	for _, tt := range tests {
		got := truncate(tt.input, tt.limit)
		if got != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.limit, got, tt.expected)
		}
	}
}
