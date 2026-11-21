package media

import (
	"context"
	"testing"

	"github.com/maccavelli/mcp-server-duckduckgo/internal/models"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestVideoPageURL(t *testing.T) {
	const watch = "https://www.youtube.com/watch?v=abc"
	tool := &MediaTool{
		Type: "videos",
		SearchFunc: func(ctx context.Context, query string, maxResults int) ([]models.SearchResult, error) {
			return []models.SearchResult{{
				URL:       watch,
				Duration:  "3:56",
				Publisher: "YouTube",
			}}, nil
		},
	}
	_, resp, err := tool.Handle(context.Background(), &mcp.CallToolRequest{}, MediaInput{
		Query:  "origami",
		Format: "json",
	})
	if err != nil {
		t.Fatal(err)
	}
	r := resp.(models.SearchResponse20)
	if len(r.Data.Results) != 1 {
		t.Fatalf("len %d", len(r.Data.Results))
	}
	if r.Data.Results[0].PageURL != watch || r.Data.Results[0].MediaURL != watch {
		t.Fatalf("page %q media %q", r.Data.Results[0].PageURL, r.Data.Results[0].MediaURL)
	}
}
