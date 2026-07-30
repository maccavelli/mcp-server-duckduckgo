package main

import (
	"testing"
)

func TestRootCmd(t *testing.T) {
	if RootCmd.Use != "mcp-server-duckduckgo" {
		t.Errorf("unexpected Use: %s", RootCmd.Use)
	}
	if RootCmd.Short == "" {
		t.Errorf("missing Short description")
	}
}
