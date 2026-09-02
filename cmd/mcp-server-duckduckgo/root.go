package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/maccavelli/mcp-server-duckduckgo/internal/config"
)

// OriginalStdout preserves the real stdout before Execute aliases it to
// stderr, so JSON-RPC output is written deliberately rather than by accident.
var OriginalStdout *os.File

// RootCmd is the command tree root. It defaults to serve when run with no
// subcommand.
var RootCmd = &cobra.Command{
	Use:   "mcp-server-duckduckgo",
	Short: "DuckDuckGo MCP Server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return serveCmd.RunE(cmd, args)
	},
}

// Execute traps stdout for protocol safety and returns the command error. It
// no longer calls os.Exit: exit mapping belongs to main so `update --check`
// can report an available update as exit 10 (MADR 0005).
func Execute() error {
	// Trap os.Stdout and safely alias it to os.Stderr to protect the JSON-RPC pipe.
	OriginalStdout = os.Stdout
	os.Stdout = os.Stderr
	RootCmd.SetOut(os.Stderr)
	RootCmd.SetErr(os.Stderr)
	RootCmd.Version = Version

	return RootCmd.Execute()
}

// loadConfig is the narrow configuration seam. Tests replace it to prove that
// update touches no configuration at all.
var loadConfig = config.Load

func init() {
	// cobra.OnInitialize(config.Load) ran for EVERY command, and config.Load
	// creates the cache directory and writes config.yaml. Merely checking for
	// an update would therefore leave files behind. A pre-run hook can be
	// scoped instead: a command annotated selfupdate.skip-config opts out.
	RootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if cmd.Annotations[skipConfigAnnotation] == skipConfigValue {
			return nil
		}
		loadConfig()
		return nil
	}
	RootCmd.AddCommand(newUpdateCmd())
}
