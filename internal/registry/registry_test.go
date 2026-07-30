package registry

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type dummyTool struct {
	name string
}

func (d *dummyTool) Name() string           { return d.name }
func (d *dummyTool) Register(s *mcp.Server) {}

func TestRegistry(t *testing.T) {
	r := &Registry{tools: make(map[string]Tool)}

	tool1 := &dummyTool{name: "tool1"}
	r.Register(tool1)

	tool, ok := r.Get("tool1")
	if !ok || tool == nil || tool.Name() != "tool1" {
		t.Error("failed to get tool1")
	}

	_, ok = r.Get("nonexistent")
	if ok {
		t.Error("expected missing tool")
	}

	list := r.List()
	if len(list) != 1 || list[0].Name() != "tool1" {
		t.Errorf("expected list to have length 1, got %v", list)
	}

	// test Global instance implicitly
	Global.Register(&dummyTool{name: "global_tool"})
	gt, ok := Global.Get("global_tool")
	if !ok || gt == nil {
		t.Error("failed to get from global registry")
	}
	if len(Global.List()) < 1 {
		t.Error("failed to list global registry")
	}
}
