# Connecting a client

Task index: [docs/README.md](../README.md).

## Production

In production the client connects to MagicTools. This binary is the downstream server MagicTools launches. Configure MagicTools in the client. This guide does not define a MagicTools config file.

When MagicTools sets `MCP_ORCHESTRATOR_OWNED=true`, `serve` does not take the singleton lock. `mcplib.IsOrchestratorOwned` is that check, and `serve.go` skips `run.lock` when it is true.

## Standalone

A standalone client launches the binary with no arguments:

```text
mcp-server-duckduckgo
```

`root.go` runs `serve` when no subcommand is given. `serve` is also a
subcommand. The transport is stdio. Protocol messages are written to the
original stdout, which `Execute` saves before it points `os.Stdout` at
stderr. Logs go to stderr.

The command value is the installed binary, on `PATH` or by its path. The installer paths are in [Installing](installing.md).

This guide does not list IDE configuration-file paths. Take the file location from that IDE's MCP documentation, and put this command in it.
