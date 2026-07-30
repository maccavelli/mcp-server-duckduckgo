package system

import (
	"github.com/maccavelli/mcplib"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/maccavelli/mcp-server-duckduckgo/internal/registry"
)

// Register registers system tools with the global registry.
func Register(lb *mcplib.LogBuffer) {
	registry.Global.Register(&diagnosticToolShim{buffer: lb})
}

type diagnosticToolShim struct {
	buffer *mcplib.LogBuffer
}

func (t *diagnosticToolShim) Name() string { return "get_internal_logs" }

func (t *diagnosticToolShim) Register(s *mcp.Server) {
	mcplib.RegisterDiagnosticTool(s, t.buffer, "duckduckgo")
}
