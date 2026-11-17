---
status: accepted
date: 2026-11-17
---
# Update dependency pins to current stable releases

## Context and Problem Statement

The owner asked to find the current versions of this module's dependencies and update as many pins as possible.

`go list -m -u all` on 2026-11-17, against the module proxy, is the version list. A line with `[vX]` is a newer release. Direct requirements already at the newest stable release are `github.com/go-rod/stealth v0.4.9`, `github.com/spf13/cobra v1.10.2`, and `github.com/spf13/viper v1.21.0`.

Direct requirements with a newer stable release:

| Module | Pinned | Current stable |
|---|---|---|
| `github.com/PuerkitoBio/goquery` | v1.11.0 | v1.14.0 |
| `github.com/maccavelli/mcp-buntdb` | v0.2.0 | v1.0.1 |
| `github.com/maccavelli/mcplib` | v1.4.1 | v1.6.0 |
| `github.com/modelcontextprotocol/go-sdk` | v1.6.1 | v1.8.0 |
| `github.com/playwright-community/playwright-go` | v0.5700.1 | v0.6401.0 |
| `github.com/tidwall/buntdb` | v1.3.2 | v1.4.0 |
| `golang.org/x/net` | v0.56.0 | v0.62.0 |
| `golang.org/x/sync` | v0.22.0 | v0.25.0 |
| `golang.org/x/sys` | v0.47.0 | v0.50.0 |

`go get -u` moves minor and patch versions. It does not cross a major version. `mcp-buntdb` v0.2.0 to v1.0.1 is that cross, on the same import path. Tag `v1.0.1` still declares `OpenBuntDB(dbPath string, config *buntdb.Config) (*buntdb.DB, error)`, which is the only call in `internal/store/store.go`.

`mcplib` v1.6.0 still exports every symbol this server calls: `HardenedAddTool`, `NewMCPServer`, `MCPServer`, `Serve`, `NewStdioPipeline`, `IsExpectedShutdownErr`, `IsOrchestratorOwned`, `NewLogBuffer`, `SetupStandardLogging`, `RegisterDiagnosticTool`, `UniversalBaseInput`, and the `selfupdate` constructors used by `update.go`. Its `go` line is 1.26.6. `go-sdk` v1.8.0 still has `NewServer`, `ServerOptions`, `Tool`, `CallToolResult.SetError`, and `TextContent`. Its `go` line is 1.25.0. `go-sdk` also publishes `v1.9.0-pre.1`. That is not the stable release `go list -m -u` selected.

Indirect modules with newer releases move when the direct upgrade, or `go get -u ./...`, selects them. They are not a second hand-written pin list.

This module's `go` line stays `1.27.4`. No candidate above declares a higher floor. `go-selfupdate-lib`, `go-llmprovider-sdk`, and `go-tui-lib` are not dependencies of this module.

Which pins move?

## Decision Drivers

* Current stable release from the module proxy, queried 2026-11-17.
* Same import path. A major move only when the called API is still that signature.
* No pre-release (`v1.9.0-pre.1`, `v1.6.0-rc1`).
* The `go 1.27.4` line stays.
* A module whose upgrade fails `go build ./...` stays on its previous pin.

## Considered Options

* Raise every direct pin that has a newer stable release, and let `go get -u ./...` raise the indirect pins those builds select
* Raise only patch releases
* Leave the pins where they are

## Decision Outcome

Chosen option: "Raise every direct pin that has a newer stable release, and let `go get -u ./...` raise the indirect pins those builds select", because the owner asked for as many current pins as possible, and the nine direct upgrades above are the ones the proxy lists. `mcp-buntdb` is requested explicitly at v1.0.1 because `-u` will not cross v0 to v1.

### Consequences

* Good, because `goquery`, `buntdb`, `x/net`, `x/sync`, and `x/sys` move inside the major this code already imports.
* Good, because `mcplib` v1.6.0 and `go-sdk` v1.8.0 still export the symbols the call sites use.
* Good, because `mcp-buntdb` v1.0.1 keeps `OpenBuntDB`'s signature.
* Neutral, because `playwright-go` v0.6100.0 and later declare `module github.com/mxschmitt/playwright-go` and cannot be required on this import path. The highest pin that resolves is v0.6000.0.
* Neutral, because indirect versions are the build list after `-u`, not a hand list.
* Bad, because a newer indirect module can change behaviour without a direct pin in the table.

### Confirmation

`go list -m` shows the nine target versions. `go` in `go.mod` is still `1.27.4`. `go build ./...` exits 0. `go list -m -u` shows no newer stable release for a direct requirement that this decision included.

## Pros and Cons of the Options

### Raise every direct pin that has a newer stable release, and let `go get -u ./...` raise the indirect pins those builds select

* Good, because it matches the proxy list and the owner's request.
* Good, because one build checks the whole set.
* Bad, because `playwright-go` v0.6401.0 may not compile against `internal/engine/playwright.go`.

### Raise only patch releases

* Good, because a patch is the smallest behaviour change.
* Bad, because `mcplib`, `go-sdk`, `goquery`, `buntdb`, `mcp-buntdb`, and `playwright-go` would stay behind the current release the owner asked for.

### Leave the pins where they are

* Good, because the tree already builds.
* Bad, because the owner asked for the current versions.

## More Information

Version query: `go list -m -u all` and `go list -m -versions` on 2026-11-17.

### Amendment 2026-11-17

`playwright-community/playwright-go` v0.6100.0 through v0.6401.0 declare `module github.com/mxschmitt/playwright-go`. The pin on this import path is v0.6000.0. `ysmood/fetchup` v0.4.0 dropped the destination argument from `New`. `go-rod` v0.116.2 still calls `New(dir, urls...)`, so fetchup is v0.3.0, the newest signature that call compiles against.

### Amendment 2026-11-17 — playwright path

The community-path pin at v0.6000.0 is superseded by [0004-MADR-move-playwright-to-maintained-module.md](0004-MADR-move-playwright-to-maintained-module.md). The fetchup pin at v0.3.0 stands.

Implementation plan: [0003-PLAN-update-dependency-pins.md](0003-PLAN-update-dependency-pins.md).

The Go floor stays the decision in [0002-MADR-pin-go-1.27.4.md](0002-MADR-pin-go-1.27.4.md). Replacing `mcplib/selfupdate` with `go-selfupdate-lib` is a separate decision and is not part of this pin update.
