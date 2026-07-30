package search

import (
	"context"
	"testing"

	"github.com/maccavelli/mcp-server-duckduckgo/internal/engine"
	"github.com/maccavelli/mcp-server-duckduckgo/internal/models"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mockSearchEngine struct{}

func (m *mockSearchEngine) WebSearch(ctx context.Context, query string, maxResults int) ([]models.SearchResult, error) {
	return []models.SearchResult{{URL: "http://example.com", Title: "Web Result"}}, nil
}
func (m *mockSearchEngine) NewsSearch(ctx context.Context, query string, maxResults int) ([]models.SearchResult, error) {
	return []models.SearchResult{{Title: "News Result"}}, nil
}
func (m *mockSearchEngine) BookSearch(ctx context.Context, query string, maxResults int) ([]models.SearchResult, error) {
	return []models.SearchResult{{Title: "Book Result"}}, nil
}
func (m *mockSearchEngine) FetchContent(ctx context.Context, rawURL string) (*engine.FetchedContent, error) {
	return &engine.FetchedContent{URL: rawURL, Title: "Mock Page", Content: "Mock content"}, nil
}

func TestSearchTool_Handle(t *testing.T) {
	eng := &mockSearchEngine{}
	tool := &SearchTool{
		Engine:     eng,
		Type:       "web",
		SearchFunc: eng.WebSearch,
		Desc:       "Test description",
	}

	ctx := context.Background()
	input := SearchInput{
		Query:      "test query",
		MaxResults: 1,
	}

	// Test Handle
	_, resp, err := tool.Handle(ctx, &mcp.CallToolRequest{}, input)
	if err != nil {
		t.Errorf("Handle failed: %v", err)
	}
	if resp == nil {
		t.Error("expected non-nil response")
	}

	// Verify Name
	if tool.Name() != "search_web" {
		t.Errorf("expected search_web, got %s", tool.Name())
	}
}

func TestRegister(t *testing.T) {
	eng := &mockSearchEngine{}
	Register(eng)

	srv := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, &mcp.ServerOptions{})
	tool := &SearchTool{Engine: eng, Type: "web"}
	tool.Register(srv)
}

func TestFetchResultContents(t *testing.T) {
	eng := &mockSearchEngine{}
	results := []models.SearchResult{
		{URL: "http://example.com/1"},
		{URL: "http://example.com/2"},
		{URL: ""},
	}
	enriched := fetchResultContents(context.Background(), eng, results)
	if len(enriched) != 3 {
		t.Fatalf("expected 3 results")
	}
	if enriched[0].Content != "Mock content" || enriched[0].Title != "Mock Page" {
		t.Errorf("expected mock content and title, got %s / %s", enriched[0].Content, enriched[0].Title)
	}
}

func TestSearchTool_Handle_FetchContent(t *testing.T) {
	eng := &mockSearchEngine{}
	tool := &SearchTool{
		Engine:     eng,
		Type:       searchTypeWeb,
		SearchFunc: eng.WebSearch,
	}
	ctx := context.Background()
	_, raw, err := tool.Handle(ctx, nil, SearchInput{
		Query:        "test",
		FetchContent: true,
		Format:       "json",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp := raw.(models.SearchResponse)
	if len(resp.Data.Results) != 1 {
		t.Fatalf("expected 1 result")
	}
	if resp.Data.Results[0].Content != "Mock content" {
		t.Errorf("expected content to be fetched")
	}
}

func TestFetchURLTool(t *testing.T) {
	eng := &mockSearchEngine{}
	tool := &FetchURLTool{Engine: eng}
	if tool.Name() != "fetch_url" {
		t.Errorf("unexpected name: %s", tool.Name())
	}
	srv := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, &mcp.ServerOptions{})
	tool.Register(srv)

	ctx := context.Background()
	res, _, err := tool.Handle(ctx, nil, FetchURLInput{URL: "http://test.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tc := res.Content[0].(*mcp.TextContent)
	if tc == nil || len(tc.Text) == 0 {
		t.Errorf("expected output, got nothing")
	}
}
