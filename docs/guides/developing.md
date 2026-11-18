# Developing

Task index: [docs/README.md](../README.md).

## Module

`go.mod` declares `go 1.26.11` and `toolchain go1.27.4`. Build and test with that toolchain. CI's `setup-go` step reads `go.mod`.

## Layout

| Path | What it holds |
| :--- | :--- |
| `cmd/mcp-server-duckduckgo` | The binary: default `serve`, plus `update` and `identity`. |
| `internal/engine` | Search and page fetch, including Open Library and the Playwright fallback. |
| `internal/handler` | MCP tool registration under `search`, `media`, and `system`. |
| `internal/store` | The BuntDB scraping store. |

The map of the running process is [architecture.md](../architecture.md).

## Commands

| Command | Behavior |
| :--- | :--- |
| `mcp-server-duckduckgo` | Runs `serve`. |
| `mcp-server-duckduckgo serve` | Same server. |
| `mcp-server-duckduckgo update` | See [Updating](updating.md). Skips configuration. |
| `mcp-server-duckduckgo identity` | Prints `buildinfo.Identity().String()` and a newline to the preserved stdout. Skips configuration. |

`identity` is what `install.sh` and `install.ps1` run after they swap the binary.

## Make targets

`make build` writes `dist/mcp-server-duckduckgo-<goos>-<goarch>` (plus `.exe` on Windows). `BUILD_KIND` defaults to `local`, so the stamp is a local build and `update` refuses to replace it without `--force`.

`make test` runs `go test -v ./...`.

`make fmt` runs `go fmt ./...` and rewrites files. `make fmt-check` lists unformatted files and does not rewrite them.

`make verify` runs, in order: `mod-check`, `fmt-check`, `lint`, `vet`, `test`, `vuln`, `check-records`, `markdownlint`, `disclosure-check`.

- `lint` runs `golangci-lint` `v2.15.0` from `$(go env GOPATH)/bin` unless `GOLANGCI_LINT` is set. The install command is `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0`.
- `vuln` runs `go run golang.org/x/vuln/cmd/govulncheck@v1.10.0 ./...`.
- `check-records` runs `python3 scripts/check_records.py --check-all`.
- `markdownlint` runs `npx --yes markdownlint-cli2@0.23.2`.
- `disclosure-check` runs `go-shared`'s `scripts/check_disclosure.py` against this repository. `GO_SHARED` defaults to `../go-shared`.

CI's validate job installs golangci-lint `v2.15.0` and runs `make verify` with `GO_SHARED` pointed at a fetched `go-shared` checkout. See [Releasing](releasing.md) for the tag path.
