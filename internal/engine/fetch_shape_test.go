package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func runExtractCellTests(t *testing.T) {
	t.Run("one_cell", func(t *testing.T) {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(`<table><tr><td><pre>package main</pre></td></tr></table>`))
		if err != nil {
			t.Fatal(err)
		}
		result := extractStructuredText(doc.Selection)
		if !strings.Contains(result, "package main") {
			t.Fatalf("missing code: %q", result)
		}
		if strings.Contains(result, "|") {
			t.Fatalf("cell separator in one-cell row: %q", result)
		}
	})

	t.Run("two_cells", func(t *testing.T) {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(`<table><tr><td>Left</td><td>Right</td></tr></table>`))
		if err != nil {
			t.Fatal(err)
		}
		result := extractStructuredText(doc.Selection)
		if !strings.Contains(result, "Left | Right") {
			t.Fatalf("missing pair: %q", result)
		}
		if strings.HasSuffix(strings.TrimSpace(result), "|") {
			t.Fatalf("trailing separator: %q", result)
		}
	})
}

func TestFetchContent_EmptyRoot(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><head><title>MSN</title></head><body><div id="root"></div></body></html>`))
	}))
	defer ts.Close()

	eng := setupMockEngine(t, ts.Client())
	res, err := eng.FetchContent(context.Background(), ts.URL)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	if res.Title != "MSN" {
		t.Fatalf("title %q", res.Title)
	}
	if res.Content != "[No readable content extracted]" {
		t.Fatalf("content %q", res.Content)
	}
}

func TestFetchContent_SkipsEmptyArticle(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><head><title>T</title></head><body><article></article><main><p>Readable text</p></main></body></html>`))
	}))
	defer ts.Close()

	eng := setupMockEngine(t, ts.Client())
	res, err := eng.FetchContent(context.Background(), ts.URL)
	if err != nil {
		t.Fatalf("err %v", err)
	}
	if !strings.Contains(res.Content, "Readable text") {
		t.Fatalf("content %q", res.Content)
	}
}
