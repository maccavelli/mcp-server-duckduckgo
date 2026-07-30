package system

import (
	"testing"

	"github.com/maccavelli/mcplib"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestRegister(t *testing.T) {
	buffer := mcplib.NewLogBuffer()
	Register(buffer)
}

func TestDiagnosticToolShim(t *testing.T) {
	buffer := mcplib.NewLogBuffer()
	shim := &diagnosticToolShim{buffer: buffer}
	if shim.Name() != "get_internal_logs" {
		t.Errorf("expected get_internal_logs, got %s", shim.Name())
	}

	srv := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0"}, &mcp.ServerOptions{})
	shim.Register(srv)
}
