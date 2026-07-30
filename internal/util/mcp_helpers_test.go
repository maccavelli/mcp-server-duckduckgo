package util

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/maccavelli/mcplib"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestHardenedAddTool(t *testing.T) {
	srv := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0"},
		&mcp.ServerOptions{},
	)
	tool := &mcp.Tool{}
	json.Unmarshal([]byte(`{"name":"test_tool","inputSchema":{"type":"object","properties":{}}}`), tool)

	handler := func(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "success"},
			},
		}, nil, nil
	}

	mcplib.HardenedAddTool(srv, tool, handler)
}
