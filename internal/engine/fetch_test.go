package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/maccavelli/mcp-server-duckduckgo/internal/store"

	"github.com/PuerkitoBio/goquery"
)

func TestExtractStructuredText(t *testing.T) {
	htmlStr := `<html><body>
		<article>
			<h1>Title</h1>
			<p>Hello <b>World</b></p>
			<ul>
				<li>One</li>
				<li>Two</li>
			</ul>
			<pre>code block</pre>
		</article>
	</body></html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlStr))
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	result := extractStructuredText(doc.Selection)

	if !strings.Contains(result, "## Title") {
		t.Errorf("expected heading: %s", result)
	}
	if !strings.Contains(result, "Hello World") {
		t.Errorf("expected text: %s", result)
	}
	if !strings.Contains(result, "- One") {
		t.Errorf("expected list: %s", result)
	}
	if !strings.Contains(result, "```\ncode block\n```") {
		t.Errorf("expected pre block: %s", result)
	}
}

func setupMockEngine(t *testing.T, client *http.Client) *SearchEngine {
	t.Helper()
	mockStore, _ := store.Open(":memory:")
	t.Cleanup(func() { mockStore.Close() })
	mockStore.SeedUAPool()
	return &SearchEngine{
		Client: client,
		store:  mockStore,
	}
}

func TestFetchContent_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><head><title>Test Title</title></head><body><article><p>Readable text</p></article></body></html>`))
	}))
	defer ts.Close()

	engine := setupMockEngine(t, ts.Client())
	ctx := context.Background()
	res, err := engine.FetchContent(ctx, ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Title != "Test Title" {
		t.Errorf("expected Title 'Test Title', got %s", res.Title)
	}
	if !strings.Contains(res.Content, "Readable text") {
		t.Errorf("expected 'Readable text' in content, got: %s", res.Content)
	}
}

func TestFetchContent_PlainText(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Just some plain text"))
	}))
	defer ts.Close()

	engine := setupMockEngine(t, ts.Client())
	res, err := engine.FetchContent(context.Background(), ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Content != "Just some plain text" {
		t.Errorf("expected plain text, got %s", res.Content)
	}
}

func TestFetchContent_Binary(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Write([]byte("fake pdf content"))
	}))
	defer ts.Close()

	engine := setupMockEngine(t, ts.Client())
	res, err := engine.FetchContent(context.Background(), ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Content, "Binary content") {
		t.Errorf("expected binary fallback, got %s", res.Content)
	}
}

func TestFetchContent_RateLimitBackoff(t *testing.T) {
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><article><p>Success after backoff</p></article></body></html>`))
	}))
	defer ts.Close()

	engine := setupMockEngine(t, ts.Client())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := engine.FetchContent(ctx, ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
	if !strings.Contains(res.Content, "Success after backoff") {
		t.Errorf("unexpected content: %s", res.Content)
	}
}

func TestFetchContent_Non200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	engine := setupMockEngine(t, ts.Client())
	res, err := engine.FetchContent(context.Background(), ts.URL)
	if err == nil {
		t.Fatal("expected error on 500")
	}
	if res.StatusCode != 500 {
		t.Errorf("expected 500 status code, got %d", res.StatusCode)
	}
}
