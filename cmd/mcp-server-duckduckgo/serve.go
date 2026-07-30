package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/maccavelli/mcplib"

	"github.com/spf13/cobra"

	"github.com/maccavelli/mcp-server-duckduckgo/internal/config"
	"github.com/maccavelli/mcp-server-duckduckgo/internal/engine"
	"github.com/maccavelli/mcp-server-duckduckgo/internal/handler/media"
	"github.com/maccavelli/mcp-server-duckduckgo/internal/handler/search"
	"github.com/maccavelli/mcp-server-duckduckgo/internal/handler/system"
	"github.com/maccavelli/mcp-server-duckduckgo/internal/lifecycle"
	"github.com/maccavelli/mcp-server-duckduckgo/internal/registry"
	"github.com/maccavelli/mcp-server-duckduckgo/internal/store"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the DuckDuckGo MCP Server",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !mcplib.IsOrchestratorOwned() {
			cacheDir, err := os.UserCacheDir()
			if err != nil {
				cacheDir = os.TempDir()
			}
			lockPath := filepath.Join(cacheDir, "mcp-server-duckduckgo", "run.lock")
			if err := os.MkdirAll(filepath.Dir(lockPath), 0o750); err != nil {
				return fmt.Errorf("failed to create lock directory: %w", err)
			}
			if err := lifecycle.TryLock(lockPath); err != nil {
				return fmt.Errorf("failed to acquire singleton lock: %w", err)
			}
			defer func() {
				if err := lifecycle.Unlock(); err != nil {
					slog.Warn("failed to release singleton lock", "error", err)
				}
			}()
		}

		if _, exists := os.LookupEnv("GOMEMLIMIT"); !exists {
			if err := os.Setenv("GOMEMLIMIT", "1024MiB"); err != nil {
				slog.Warn("failed to set GOMEMLIMIT", "error", err)
			}
		}
		if _, exists := os.LookupEnv("GOMAXPROCS"); !exists {
			if err := os.Setenv("GOMAXPROCS", "2"); err != nil {
				slog.Warn("failed to set GOMAXPROCS", "error", err)
			}
		}

		logBuffer := mcplib.NewLogBuffer()
		cleanupLogs := mcplib.SetupStandardLogging("duckduckgo", logBuffer)
		defer cleanupLogs()

		slog.Info("[BACKPLANE] SPAWN "+config.Name, "version", cmd.Version)

		rootCtx := context.Background()
		ctx, stop := signal.NotifyContext(rootCtx, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
		defer stop()

		pipeline := mcplib.NewStdioPipeline(os.Stdin, OriginalStdout, stop)

		if err := run(ctx, logBuffer, pipeline.Reader, pipeline.Writer, cmd.Version); err != nil {
			if mcplib.IsExpectedShutdownErr(err) {
				slog.Info("server shut down gracefully", "error", err)
				if flushErr := pipeline.Flush(); flushErr != nil {
					slog.Warn("failed to flush stdio pipeline", "error", flushErr)
				}
				return nil
			}
			slog.Error("server fatal error", "error", err)
			return err
		}
		if flushErr := pipeline.Flush(); flushErr != nil {
			slog.Warn("failed to flush stdio pipeline", "error", flushErr)
		}
		return nil
	},
}

func run(ctx context.Context, lb *mcplib.LogBuffer, reader io.ReadCloser, writer io.WriteCloser, version string) error {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}
	dataDir := filepath.Join(cacheDir, "mcp-server-duckduckgo")

	if err := os.MkdirAll(dataDir, 0o750); err != nil {
		return fmt.Errorf("failed to create data dir: %w", err)
	}

	scrapingStore, err := store.Open(dataDir)
	if err != nil {
		return fmt.Errorf("failed to open scraping store: %w", err)
	}
	// STD-MCP-CLI-LIFECYCLE-001: defer Close() immediately after initialization.
	defer func() {
		if err := scrapingStore.Close(); err != nil {
			slog.Warn("failed to close scraping store", "error", err)
		}
	}()

	// Seed the UA pool on first run.
	scrapingStore.SeedUAPool()

	eng := engine.NewSearchEngine(scrapingStore)
	search.Register(eng)
	media.Register(eng)
	system.Register(lb)

	mcpServer := mcplib.NewMCPServer(config.Name, version, slog.Default())

	for _, t := range registry.Global.List() {
		t.Register(mcpServer.MCPServer())
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- mcpServer.Serve(ctx, writer, reader)
	}()

	select {
	case <-ctx.Done():
		slog.Info("context cancelled; initiating graceful shutdown")
	case err := <-errChan:
		if err == nil {
			return nil
		}
		if mcplib.IsExpectedShutdownErr(err) {
			slog.Info("stdio transport closed gracefully", "reason", err.Error())
			return nil
		}
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

func init() {
	RootCmd.AddCommand(serveCmd)
}
