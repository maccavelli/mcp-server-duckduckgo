package search

import (
	"context"
	"errors"
	"testing"

	"github.com/maccavelli/mcp-server-duckduckgo/internal/engine"
	"github.com/maccavelli/mcp-server-duckduckgo/internal/models"
)

type errFetchEngine struct{ mockEngine }

func (e *errFetchEngine) FetchContent(ctx context.Context, rawURL string) (*engine.FetchedContent, error) {
	return nil, errors.New("fetch failed with status 503")
}

func TestFetchResultContents_FetchError(t *testing.T) {
	results := []models.SearchResult{{URL: "http://example.com/down", Title: "Keep"}}
	enriched := fetchResultContents(context.Background(), &errFetchEngine{}, results)
	if len(enriched) != 1 {
		t.Fatalf("len %d", len(enriched))
	}
	if enriched[0].Title != "Keep" {
		t.Fatalf("hit dropped: %+v", enriched[0])
	}
	if enriched[0].Content != "fetch failed with status 503" {
		t.Fatalf("content %q", enriched[0].Content)
	}
}

type markerFetchEngine struct{ mockEngine }

func (e *markerFetchEngine) FetchContent(ctx context.Context, rawURL string) (*engine.FetchedContent, error) {
	return &engine.FetchedContent{Content: "[No readable content extracted]"}, nil
}

func TestFetchResultContents_EmptyMarker(t *testing.T) {
	results := []models.SearchResult{{URL: "http://example.com/msn"}}
	enriched := fetchResultContents(context.Background(), &markerFetchEngine{}, results)
	if len(enriched) != 1 || enriched[0].Content != "[No readable content extracted]" {
		t.Fatalf("got %+v", enriched)
	}
}
