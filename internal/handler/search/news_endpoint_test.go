package search

import (
	"context"
	"strings"
	"testing"

	"github.com/maccavelli/mcp-server-duckduckgo/internal/models"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestNewsToolDesc(t *testing.T) {
	if !strings.Contains(newsToolDesc, "short keyword query") {
		t.Fatal("news tool description does not name a short keyword query")
	}
}

func TestNewsEndpoint(t *testing.T) {
	eng := &mockEngine{}
	tool := &Tool{
		Engine:     eng,
		Type:       "news",
		SearchFunc: eng.NewsSearch,
	}
	_, resp, err := tool.Handle(context.Background(), &mcp.CallToolRequest{}, Input{
		Query:  "Federal Reserve",
		Format: "json",
	})
	if err != nil {
		t.Fatal(err)
	}
	r := resp.(models.SearchResponse)
	if r.Data.Metadata.Endpoint != "DuckDuckGo" {
		t.Fatalf("endpoint %q", r.Data.Metadata.Endpoint)
	}
}
