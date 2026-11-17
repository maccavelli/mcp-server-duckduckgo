---
status: implemented
date: 2026-11-17
associated-madr: "0005-MADR-pin-golangci-lint-v2.15.0.md"
---
# Implement the golangci-lint v2.15.0 pin

Associated MADR: [0005-MADR-pin-golangci-lint-v2.15.0.md](0005-MADR-pin-golangci-lint-v2.15.0.md)

## Goal

Make `make lint` pass on Go 1.27.4 by installing `golangci-lint` v2.15.0 and clearing the three findings that binary reports.

## Scope

In scope:

* The `go install` version in `.github/workflows/ci.yml`
* The `olBareWorkKey` pattern in `internal/engine/books.go`
* The unused `errorReader` type in `internal/engine/test_helper_test.go`, and the `fmt` import if nothing else in that file needs it
* `docs/README.md` index row for this pair

Out of scope:

* The `go 1.27.4` line and `Makefile` `MOD_VERSION`
* GitHub Actions SHAs
* `.golangci.yml` (enable list, exclusions, and issue limits stay)
* `go test ./...` (heavy; this plan does not name a fleet host)
* Commit and push
* The `Makefile` comment that cites mcplib's build-kind record

## Grounding Reference

[0005-MADR-pin-golangci-lint-v2.15.0.md](0005-MADR-pin-golangci-lint-v2.15.0.md) records why golangci-lint v2.13.1 cannot decode Go 1.27.4 export data and why v2.15.0 can. Before the edits, v2.15.0 reported three findings on this tree: `regexpSimplify` on `olBareWorkKey`, the unused `errorReader` type, and the unused `fmt` import.

## Prerequisites

* `go.mod` says `go 1.27.4`.
* `.golangci.yml` stays as it is. Do not disable a linter to clear a finding.
* `make lint` runs `golangci-lint run -c .golangci.yml ./...`.

## Implementation Steps

### Phase 1 — Pin the linter and clear the three findings

From the repository root, on the session host:

1. In `.github/workflows/ci.yml`, change `golangci-lint@v2.13.1` to `golangci-lint@v2.15.0`.
2. In `internal/engine/books.go`, change the `olBareWorkKey` pattern from `` `^OL[0-9]+W$` `` to `` `^OL\d+W$` ``. In Go's `regexp` package, `\d` is `[0-9]`.
3. Delete `errorReader` and its `Read` method from `internal/engine/test_helper_test.go`. Drop the `fmt` import when it is unused.
4. `gofmt` both Go files.

Stop if `golangci-lint` v2.15.0 reports any finding other than the three named in the MADR before those edits, or any finding after them. Record that output. Do not edit `.golangci.yml` to hide a finding. Do not leave the workflow pin on v2.13.1.

## Verification

Install the binary without replacing the user's `GOBIN` entry, then run the same config CI uses:

```bash
GOBIN=/tmp/gcl-bin go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0
/tmp/gcl-bin/golangci-lint version
GOLANGCI_LINT=/tmp/gcl-bin/golangci-lint make lint
go test -count=1 ./internal/engine -run TestBookSearch
```

`golangci-lint version` prints 2.15.0. `make lint` exits 0. `TestBookSearch` exits 0, including `TestBookSearchMapsWorks`, which maps the bare key `OL2W`. The `go` line in `go.mod` is still 1.27.4. The workflow installs `@v2.15.0`.

## Rollout and Rollback

Rollback is restoring `@v2.13.1`, the character class `` `[0-9]` ``, and `errorReader`. That rollback returns CI to the export-data failure on Go 1.27.4.

## Deviations

None yet.

## Execution record

### 2026-11-17 — pin applied

`.github/workflows/ci.yml` installs `golangci-lint@v2.15.0`. `olBareWorkKey` is `` `^OL\d+W$` ``. `errorReader` is gone from `internal/engine/test_helper_test.go`, and that file no longer imports `fmt`. `gofmt -l` printed nothing for the two Go files. `golint` on those files exited 0.

`GOBIN=/tmp/gcl-bin go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0` exited 0. `golangci-lint version` printed 2.15.0 built with go1.27.4. `GOLANGCI_LINT=/tmp/gcl-bin/golangci-lint make lint` exited 0 and printed `0 issues.` `go test -count=1 ./internal/engine -run TestBookSearch` exited 0 (`ok`, 0.775s). `go.mod` still says `go 1.27.4`. No commit.

### 2026-11-17 — record profile

Frontmatter `status` is `implemented`. `check_records.py` rejects `complete`. `## Grounding Reference`, `## Prerequisites`, `## Rollback`, and `## Completion Criteria` were added for the strict record profile. The phase text and the pin notes above are unchanged.

## Rollback

The steps are under Rollout and Rollback. Restore `@v2.13.1`, the character class `` `[0-9]` ``, and `errorReader`. That rollback returns CI to the export-data failure on Go 1.27.4.

## Completion Criteria

* `.github/workflows/ci.yml` installs `golangci-lint@v2.15.0`.
* `olBareWorkKey` is `` `^OL\d+W$` ``. `errorReader` and the unused `fmt` import are gone.
* `make lint` with v2.15.0 exited 0. `TestBookSearch` exited 0.
* `go.mod` still says `go 1.27.4`.
* Frontmatter `status` is `implemented`.
