---
status: implemented
date: 2026-11-17
associated-madr: "0002-MADR-pin-go-1.27.4.md"
---
# Implement the Go 1.27.4 pin and rebuild

Associated MADR: [0002-MADR-pin-go-1.27.4.md](0002-MADR-pin-go-1.27.4.md)

## Goal

Record Go 1.27.4 as this module's `go` line and as `MOD_VERSION`, then rebuild the current working tree into `dist/` with the session host's `go1.27.4`.

## Scope

In scope:

* `go.mod` `go` line, via `go get go@1.27.4`
* `Makefile` `MOD_VERSION`
* `make build-all` (linux/amd64, darwin/arm64, windows/amd64) into gitignored `dist/`
* `docs/README.md` index row for this pair (written with the pair)

Out of scope:

* An explicit `toolchain` line, a `GOTOOLCHAIN=` prefix in recipes, or a `go-version` input in `.github/workflows/ci.yml`
* `VERSION`, product release metadata, and dependency bumps
* `go mod tidy` as its own step
* Re-running [0001-PLAN-retarget-book-search-open-library.md](0001-PLAN-retarget-book-search-open-library.md)
* `make test`, `go test ./...`, `-race`, or fuzz. Those are heavy runs and this plan does not name a fleet host for them. The book-search packages were already tested on this same `go1.27.4`.
* Commit and push. HEAD is `main`. End the phase with a dirty tree.

## Grounding Reference

The version pair, and the reasons for omitting a `toolchain` line, are in [0002-MADR-pin-go-1.27.4.md](0002-MADR-pin-go-1.27.4.md). Phase 1 runs on the session host that already reports Go 1.27.4, as Implementation Steps states. Commit `cd66ef28` is the pair this plan copies.

## Prerequisites

* The session host's `go` is 1.27.4 and `GOTOOLCHAIN` is `go1.27.4`.
* Open Library edits already in the tree are pre-existing and are not this phase's diff.
* `make build-all` is three `go build` invocations and stays on the session host.

## Implementation Steps

### Phase 1 — Set the floor

Run from the repository root on the session host. That host already reports `go version go1.27.4 darwin/arm64` and `GOTOOLCHAIN=go1.27.4`.

1. Run `go get go@1.27.4`.
2. Set `MOD_VERSION := 1.27.4` in `Makefile`. The recipes do not expand it; the assignment stays paired with the `go` line, as in `cd66ef28`.
3. Stop if any of these fail. Do not tidy them away:
   * ~~`go.mod` contains a line other than the `go` directive in its diff~~ Waived 2026-11-17. See deviation below. The require regroup stays. A version change, a new module, or a dropped module still stops the phase.
   * a `toolchain` line was added
   * `go.sum` differs
   * any file outside `go.mod` and `Makefile` changed because of this phase

The Open Library edits already in the tree are pre-existing. They are not part of this phase's diff.

### Phase 2 — Rebuild `dist/`

`make build-all` is three `go build` invocations, not a suite test. Run it on the session host. Do not dispatch it with `run-on-host`.

1. Run `make build-all`.
2. The recipes stay as they are. `linux` sets `CGO_ENABLED=0`. `darwin-arm64` and `windows-amd64` do not. A cross-compile failure stops the phase.

The binaries include the uncommitted Open Library book-search sources because `go build` compiles the working tree.

## Verification

After phase 1:

* `go.mod` begins with `module github.com/maccavelli/mcp-server-duckduckgo` and `go 1.27.4`, and has no `toolchain` line.
* `Makefile` line 1 is `MOD_VERSION := 1.27.4`.
* `git diff -- go.sum` is empty.
* `go env GOVERSION` prints `go1.27.4`.

After phase 2, each of these exits 0 and the first line contains `go1.27.4`:

* `go version -m dist/mcp-server-duckduckgo-darwin-arm64`
* `go version -m dist/mcp-server-duckduckgo-linux-amd64`
* `go version -m dist/mcp-server-duckduckgo-windows-amd64.exe`

Do not execute the linux or windows binaries on this host.

## Rollout and Rollback

`dist/` is gitignored. Replacing those files affects only the local workspace. CI picks up the new floor on the next push that contains `go.mod`, because both jobs use `go-version-file: go.mod`.

Rollback is restoring `go 1.26.6` and `MOD_VERSION := 1.26.6`. Rebuild `dist/` after that restore if the binaries should match the restored floor.

## Deviations

### 2026-11-17 — Keep the `go get` require regroup

`go get go@1.27.4` changed `go 1.26.6` to `go 1.27.4` and regrouped `require` blocks. `go.sum` did not change. The 49 require lines are the same multiset as `HEAD` (path, version, and `// indirect`). No `toolchain` line. The blocks that moved:

* `github.com/maccavelli/mcp-buntdb v0.2.0` and `github.com/maccavelli/mcplib v1.4.1` into the first `require` block
* `golang.org/x/mod v0.40.0 // indirect` and `golang.org/x/term v0.44.0 // indirect` into the other indirect block

The owner chose to keep that layout. Phase 1 continues with the `go.mod` `go get` wrote. No files were added to the phase. `Makefile` is still the only other edit.

## Execution record

### 2026-11-16 — Phase 1 stopped after `go get`

`go get go@1.27.4` exited 0 and printed `go: upgraded go 1.26.6 => 1.27.4`. `go.sum` is unchanged. No `toolchain` line. The require-line multiset is identical to `HEAD` (49 lines, same path, version, and `// indirect` comments). `go.mod` also regrouped blocks: `github.com/maccavelli/mcp-buntdb v0.2.0` and `github.com/maccavelli/mcplib v1.4.1` moved into the first `require` block, and `golang.org/x/mod v0.40.0 // indirect` and `golang.org/x/term v0.44.0 // indirect` joined the remaining indirect block.

The plan says to stop when the `go.mod` diff contains a line other than the `go` directive. Phase 1 step 2 and phase 2 had not run yet. Resolved 2026-11-17 by the deviation above.

### 2026-11-17 — Phase 1 finished and phase 2 rebuilt `dist/`

`Makefile` line 1 is `MOD_VERSION := 1.27.4`. `go.mod` begins:

```text
module github.com/maccavelli/mcp-server-duckduckgo

go 1.27.4
```

No `toolchain` line. `git diff --exit-code -- go.sum` exited 0. `go env GOVERSION` printed `go1.27.4`.

`make build-all` exited 0. The recipes stamped `RawVersion=v1.1.0-dirty` and `RawBuildKind=local` because the working tree is dirty. `go version -m` exited 0 for each binary. First lines:

```text
dist/mcp-server-duckduckgo-darwin-arm64: go1.27.4
dist/mcp-server-duckduckgo-linux-amd64: go1.27.4
dist/mcp-server-duckduckgo-windows-amd64.exe: go1.27.4
```

The linux and windows binaries were not executed. No commit.

### 2026-11-17 — record profile

Frontmatter `status` is `implemented`. `check_records.py` rejects `complete`. `## Grounding Reference`, `## Prerequisites`, `## Rollback`, and `## Completion Criteria` were added for the strict record profile. The two fences above are tagged `text`. The phase text and the earlier execution notes are unchanged.

## Rollback

The steps are under Rollout and Rollback. Restore `go 1.26.6` and `MOD_VERSION := 1.26.6`, then rebuild `dist/` if the binaries should match that floor.

## Completion Criteria

* `go.mod` says `go 1.27.4` and has no `toolchain` line.
* `Makefile` line 1 is `MOD_VERSION := 1.27.4`.
* `go.sum` is unchanged.
* `go version -m` on each `dist/` binary reports `go1.27.4`.
* Frontmatter `status` is `implemented`.

## Amendment 2026-11-17 (second)

The `go` line this plan set to 1.27.4 became `go 1.26.11`, with `toolchain go1.27.4`, under [0007-PLAN-move-toolchain-to-go-1-27-4-and-floor-to-1-26-11.md](0007-PLAN-move-toolchain-to-go-1-27-4-and-floor-to-1-26-11.md). The steps and the execution record above are history and are unchanged.
