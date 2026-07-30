package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/maccavelli/mcp-server-duckduckgo/internal/config"
)

var OriginalStdout *os.File

var RootCmd = &cobra.Command{
	Use:   "mcp-server-duckduckgo",
	Short: "DuckDuckGo MCP Server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return serveCmd.RunE(cmd, args)
	},
}

func Execute() {
	// Trap os.Stdout and safely alias it to os.Stderr to protect the JSON-RPC pipe.
	OriginalStdout = os.Stdout
	os.Stdout = os.Stderr
	RootCmd.SetOut(os.Stderr)
	RootCmd.SetErr(os.Stderr)

	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(config.Load)
}
