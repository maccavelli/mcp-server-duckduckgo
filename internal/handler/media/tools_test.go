package media

import (
	"context"
	"strings"
	"testing"

	"github.com/maccavelli/mcp-server-duckduckgo/internal/models"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mockMediaEngine struct{}

func (m *mockMediaEngine) ImageSearch(ctx context.Context, query string, maxResults int) ([]models.SearchResult, error) {
	return []models.SearchResult{{Title: "Image Result", ImageURL: "http://img.com/1"}}, nil
}
func (m *mockMediaEngine) VideoSearch(ctx context.Context, query string, maxResults int) ([]models.SearchResult, error) {
	return []models.SearchResult{{Title: "Video Result", URL: "http://vid.com/1"}}, nil
}

func TestMediaTool_Handle(t *testing.T) {
	engine := &mockMediaEngine{}
	tool := &MediaTool{
		Engine:     engine,
		Type:       "images",
		SearchFunc: engine.ImageSearch,
		Desc:       "Test description",
	}

	ctx := context.Background()

	t.Run("default format (json)", func(t *testing.T) {
		input := MediaInput{
			Query:      "test image",
			MaxResults: 1,
		}
		_, resp, err := tool.Handle(ctx, &mcp.CallToolRequest{}, input)
		if err != nil {
			t.Errorf("Handle failed: %v", err)
		}
		r := resp.(models.SearchResponse20)
		if len(r.Data.Results) != 1 {
			t.Errorf("expected 1 result")
		}
	})

	t.Run("markdown format", func(t *testing.T) {
		input := MediaInput{
			Query:      "test image",
			MaxResults: 1,
			Format:     "markdown",
		}
		res, _, err := tool.Handle(ctx, &mcp.CallToolRequest{}, input)
		if err != nil {
			t.Errorf("Handle failed: %v", err)
		}
		tc := res.Content[0].(*mcp.TextContent)
		if !strings.Contains(tc.Text, "Image Result") {
			t.Errorf("expected text content, got %s", tc.Text)
		}
	})

	t.Run("hybrid format", func(t *testing.T) {
		input := MediaInput{
			Query:      "test video",
			MaxResults: 1,
			Format:     "hybrid",
		}

		toolVid := &MediaTool{
			Engine:     engine,
			Type:       "videos",
			SearchFunc: engine.VideoSearch,
		}
		_, resp, err := toolVid.Handle(ctx, &mcp.CallToolRequest{}, input)
		if err != nil {
			t.Errorf("Handle failed: %v", err)
		}
		r := resp.(models.SearchResponse20)
		if !strings.Contains(r.ResultsMD, "Video Result") {
			t.Errorf("expected hybrid MD content")
		}
	})

	// Verify Name
	if tool.Name() != "search_images" {
		t.Errorf("expected search_images, got %s", tool.Name())
	}
}

func TestRegister(t *testing.T) {
	eng := &mockMediaEngine{}
	Register(eng)

	srv := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, &mcp.ServerOptions{})
	tool := &MediaTool{Engine: eng, Type: "images"}
	tool.Register(srv)
}
