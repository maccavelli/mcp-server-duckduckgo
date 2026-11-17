---
status: implemented
date: 2026-11-17
associated-madr: "0004-MADR-move-playwright-to-maintained-module.md"
---
# Implement the playwright module-path move

Associated MADR: [0004-MADR-move-playwright-to-maintained-module.md](0004-MADR-move-playwright-to-maintained-module.md)

## Goal

Require `github.com/mxschmitt/playwright-go` v0.6401.0 and point the single import at that module.

## Scope

In scope:

* The import in `internal/engine/playwright.go`
* `go.mod` and `go.sum` for that module change, including an indirect module that `go mod tidy` drops because only community v0.6000.0 required it
* `docs/README.md` index row for this pair
* A pointer in [0003-MADR-update-dependency-pins.md](0003-MADR-update-dependency-pins.md) that this decision supersedes the community playwright pin

Out of scope:

* Stealth script injection, cookie field copies, SameSite handling, and the five-second wait
* The `go 1.27.4` line and `Makefile` `MOD_VERSION`
* Raising `github.com/ysmood/fetchup` above v0.3.0
* Adding `go-selfupdate-lib`, `go-llmprovider-sdk`, or `go-tui-lib`
* Pre-releases
* GitHub Actions SHAs
* `go test ./...` (heavy; this plan does not name a fleet host)
* Commit and push
* Running `playwright.Install()` or downloading browsers in this phase

## Grounding Reference

[0004-MADR-move-playwright-to-maintained-module.md](0004-MADR-move-playwright-to-maintained-module.md) requires `github.com/mxschmitt/playwright-go` v0.6401.0. That tag's `go.mod` declares the mxschmitt module path. [0003-PLAN-update-dependency-pins.md](0003-PLAN-update-dependency-pins.md) left the community module at v0.6000.0 because `go get` of v0.6401.0 on the community path failed.

## Prerequisites

* The pin update from 0003 is already in the tree. Leave it in place.
* `github.com/ysmood/fetchup` is v0.3.0. If `go list -m` shows another version, pin v0.3.0 again.
* `go 1.27.4` stays. This phase does not run `playwright.Install()`.

## Implementation Steps

### Phase 1 — Switch the module

From the repository root, on the session host:

1. Replace the import `github.com/playwright-community/playwright-go` in `internal/engine/playwright.go` with `github.com/mxschmitt/playwright-go`.
2. Run:

   ```bash
   go get github.com/mxschmitt/playwright-go@v0.6401.0
   go mod tidy
   ```

3. If `go list -m github.com/ysmood/fetchup` is anything other than v0.3.0, run `go get github.com/ysmood/fetchup@v0.3.0` and `go mod tidy` again.
4. `gofmt` `internal/engine/playwright.go`.

Stop if `go.mod` no longer says `go 1.27.4`. Stop if a selected version is a pre-release (`-pre` or `-rc` before the end of the version). If `go build ./...` fails, restore the community import and the `go.mod` / `go.sum` from before this phase, record the deviation, and build again. Do not leave a failing tree. Leave the uncommitted pin update from [0003-PLAN-update-dependency-pins.md](0003-PLAN-update-dependency-pins.md) in place.

## Verification

* `go list -m github.com/mxschmitt/playwright-go` prints v0.6401.0.
* `go list -m github.com/playwright-community/playwright-go` reports that the community module is no longer required.
* `go list -m github.com/ysmood/fetchup` prints v0.3.0.
* The first lines of `go.mod` contain `go 1.27.4` and no `toolchain` line.
* `go build ./...` exits 0 on the session host.
* `go test -count=1 -run '^$'` for `./internal/engine`, `./internal/store`, `./internal/handler/search`, `./internal/handler/media`, `./internal/handler/system`, `./internal/util`, and `./cmd/mcp-server-duckduckgo` exits 0. That compiles the test files and runs no tests. It is not `go test ./...`.

## Rollout and Rollback

Rollback is restoring the community import and the `go.mod` and `go.sum` saved at the start of the phase. A machine that already downloaded the 1.64.1 driver keeps that cache; rollback of the module pin does not delete it.

## Deviations

None. `go.mongodb.org/mongo-driver` stayed because `golang-set/v2` imports it. That matches the tidy rule in the scope: drop an indirect module only when nothing else requires it.

## Execution record

### 2026-11-17 — module switched

The import in `internal/engine/playwright.go` is `github.com/mxschmitt/playwright-go`. `gofmt -l` printed nothing for that file.

`go get github.com/mxschmitt/playwright-go@v0.6401.0` exited 0. `go mod tidy` exited 0.

`go.mod` still says `go 1.27.4`. No `toolchain` line. `github.com/ysmood/fetchup` is v0.3.0. The community module is no longer a dependency. `go list -m github.com/mxschmitt/playwright-go` prints v0.6401.0.

`go.sum` added `github.com/mxschmitt/playwright-go` and dropped `github.com/playwright-community/playwright-go`, `github.com/go-jose/go-jose/v3`, `github.com/stretchr/objx`, `github.com/yuin/goldmark`, `golang.org/x/crypto`, and `golang.org/x/xerrors`.

`go list -m all` also prints `go.yaml.in/yaml/v4 v4.0.0-rc.2`. No package in `go list -deps ./...` imports it. It is a requirement inside `mcplib` v1.6.0's `go.mod`, and it is absent from this module's `go.mod` and `go.sum`. That mcplib pin did not change in this phase.

`go build ./...` exited 0. `go test -count=1 -run '^$'` exited 0 for `./internal/engine`, `./internal/store`, `./internal/handler/search`, `./internal/handler/media`, `./internal/handler/system`, `./internal/util`, and `./cmd/mcp-server-duckduckgo`. Each line was `ok` and `[no tests to run]`. No commit.

### 2026-11-17 — record profile

Frontmatter `status` is `implemented`. `check_records.py` rejects `complete`. `## Grounding Reference`, `## Prerequisites`, `## Rollback`, and `## Completion Criteria` were added for the strict record profile. The `go get` fence is indented under step 2 so the ordered list stays 1 through 4. The phase text and the module-switch notes above are unchanged.

## Rollback

The steps are under Rollout and Rollback. Restore the community import and the `go.mod` and `go.sum` saved at the start of the phase. A machine that already downloaded the 1.64.1 driver keeps that cache. Rolling back the module pin does not delete it.

## Completion Criteria

* `go list -m github.com/mxschmitt/playwright-go` prints v0.6401.0.
* The community module is no longer required. `github.com/ysmood/fetchup` is v0.3.0.
* `go.mod` still says `go 1.27.4` and has no `toolchain` line.
* `go build ./...` and the compile-only package tests exited 0.
* Frontmatter `status` is `implemented`.
