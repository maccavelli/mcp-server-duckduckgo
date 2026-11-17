---
status: complete
date: 2026-11-17
associated-madr: "0003-MADR-update-dependency-pins.md"
---
# Implement the dependency pin update

Associated MADR: [0003-MADR-update-dependency-pins.md](0003-MADR-update-dependency-pins.md)

## Goal

Pin the nine direct modules in the MADR to the stable releases listed there, and upgrade every other selected module to its latest minor or patch.

## Scope

In scope:

* `go.mod` and `go.sum`
* `docs/README.md` index row for this pair

Out of scope:

* The `go 1.27.4` line and `Makefile` `MOD_VERSION`
* Adding `go-selfupdate-lib`, `go-llmprovider-sdk`, or `go-tui-lib`
* Pre-releases
* GitHub Actions SHAs
* `go test ./...` (heavy; this plan does not name a fleet host)
* Commit and push

## Implementation Steps

### Phase 1 — Set the pins

From the repository root, on the session host:

```bash
go get github.com/PuerkitoBio/goquery@v1.14.0 \
  github.com/maccavelli/mcp-buntdb@v1.0.1 \
  github.com/maccavelli/mcplib@v1.6.0 \
  github.com/modelcontextprotocol/go-sdk@v1.8.0 \
  github.com/playwright-community/playwright-go@v0.6401.0 \
  github.com/tidwall/buntdb@v1.4.0 \
  golang.org/x/net@v0.62.0 \
  golang.org/x/sync@v0.25.0 \
  golang.org/x/sys@v0.50.0
go get -u ./...
go mod tidy
```

`mcp-buntdb@v1.0.1` is explicit because `go get -u` does not cross v0 to v1.

Stop if `go.mod` no longer says `go 1.27.4`. Stop if a selected version is a pre-release. If `go build ./...` fails, return the module that broke to its previous version, record the deviation, and build again. Do not leave a failing tree.

## Verification

* `go list -m` prints the nine target versions.
* The first lines of `go.mod` contain `go 1.27.4` and no `toolchain` line.
* `go build ./...` exits 0 on the session host.
* `go test -count=1 -run '^$'` for `./internal/store`, `./internal/handler/search`, `./internal/handler/media`, `./internal/handler/system`, `./internal/util`, and `./cmd/mcp-server-duckduckgo` exits 0. That compiles the test files and runs no tests. It is not `go test ./...`.

## Rollout and Rollback

Rollback is restoring the previous `go.mod` and `go.sum` for any module this phase moved.

## Deviations

### 2026-11-17 — playwright-go stays on the community module path

`go get github.com/playwright-community/playwright-go@v0.6401.0` failed. The tag's `go.mod` declares `module github.com/mxschmitt/playwright-go`. The same mismatch is on v0.6201.1, v0.6201.0, and v0.6100.0. `go get -u ./...` then selected v0.6000.0, the newest tag that still declares `github.com/playwright-community/playwright-go`. The import path in this repo stays. v0.5700.1 moves to v0.6000.0.

### 2026-11-17 — fetchup stays at v0.3.0

`go get -u` selected `github.com/ysmood/fetchup` v0.5.3 and `github.com/go-rod/rod` v0.116.2. `go build ./...` failed in `rod@v0.116.2/lib/launcher/browser.go:138`: `fetchup.New(dir, us...)` does not match `func New(us ...string)`. v0.3.0 is the newest fetchup whose `New` is still `func New(to string, us ...string)`. rod v0.116.2 stays. fetchup is pinned to v0.3.0.

## Execution record

### 2026-11-17 — pins applied

`go 1.27.4` is unchanged. No `toolchain` line. No pre-release was selected.

Direct results:

| Module | Result |
|---|---|
| `PuerkitoBio/goquery` | v1.14.0 |
| `maccavelli/mcp-buntdb` | v1.0.1 |
| `maccavelli/mcplib` | v1.6.0 |
| `modelcontextprotocol/go-sdk` | v1.8.0 |
| `playwright-community/playwright-go` | v0.6000.0 |
| `tidwall/buntdb` | v1.4.0 |
| `golang.org/x/net` | v0.62.0 |
| `golang.org/x/sync` | v0.25.0 |
| `golang.org/x/sys` | v0.50.0 |
| `go-rod/stealth` | v0.4.9, already current |
| `spf13/cobra` | v1.10.2, already current |
| `spf13/viper` | v1.21.0, already current |

`go build ./...` exited 0 after the fetchup pin. `go test -count=1 -run '^$'` exited 0 for `./internal/store`, `./internal/handler/search`, `./internal/handler/media`, `./internal/handler/system`, `./internal/util`, and `./cmd/mcp-server-duckduckgo`. No commit.
