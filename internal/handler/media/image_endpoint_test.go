package media

import (
	"context"
	"testing"

	"github.com/maccavelli/mcp-server-duckduckgo/internal/models"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestImageEndpointBingSource(t *testing.T) {
	tool := &MediaTool{
		Type: "images",
		SearchFunc: func(ctx context.Context, query string, maxResults int) ([]models.SearchResult, error) {
			return []models.SearchResult{{
				Title:    "Hubble",
				URL:      "https://example.com/page",
				ImageURL: "https://tse.mm.bing.net/th",
				Source:   "Bing",
			}}, nil
		},
	}
	_, resp, err := tool.Handle(context.Background(), &mcp.CallToolRequest{}, MediaInput{
		Query:  "hubble",
		Format: "json",
	})
	if err != nil {
		t.Fatal(err)
	}
	r := resp.(models.SearchResponse20)
	if r.Data.Results[0].Source != "Bing" || r.Data.Metadata.Endpoint != "DuckDuckGo" {
		t.Fatalf("source %q endpoint %q", r.Data.Results[0].Source, r.Data.Metadata.Endpoint)
	}
}
