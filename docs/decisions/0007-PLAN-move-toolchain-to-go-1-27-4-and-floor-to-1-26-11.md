---
status: in-progress
date: 2026-11-17
associated-madr: "0007-MADR-move-toolchain-to-go-1-27-4-and-floor-to-1-26-11.md"
---
# Implement `go 1.26.11` as the floor and `toolchain go1.27.4`

Associated MADR: [0007-MADR-move-toolchain-to-go-1-27-4-and-floor-to-1-26-11.md](0007-MADR-move-toolchain-to-go-1-27-4-and-floor-to-1-26-11.md)

## Goal

* `go.mod` says `go 1.26.11` and `toolchain go1.27.4`, and every current statement of a Go version in the tree agrees.
* CI keeps building with 1.27.4, and the floor builds, vets and tests under 1.26.11.
* The plan stops before any tag; the tag and the release need the owner's word.

## Scope

In scope:

* `go.mod:3` and the new `toolchain` line.
* `Makefile:1`.
* `docs/README.md` rows.
* A dated amendment note in `0002-MADR-pin-go-1.27.4.md` and `0002-PLAN-pin-go-1.27.4.md`.
* The gates below.

Out of scope:

* `ci.yml` (the lint pin is already v2.15.0 and both jobs follow `go.mod`), any new CI job, `go.sum`, any dependency, `VERSION`, and any tag or release.
* Fixing or suppressing any lint finding (there are none at v2.15.0; any would go to the owner).
* Commit and push: the coordinator commits the records, then the change, on the owner's approval of 2026-11-17.

## Grounding Reference

The decision, evidence index E1 to E9 and the not-verified list are in [0007-MADR-move-toolchain-to-go-1-27-4-and-floor-to-1-26-11.md](0007-MADR-move-toolchain-to-go-1-27-4-and-floor-to-1-26-11.md). The model is the longer pair in go-selfupdate-lib, `0045-PLAN-move-toolchain-to-go-1-27-4-and-floor-to-1-26-11.md`.

## Prerequisites

* HEAD `7acaf33` on `main`, clean, level with `origin/main`.
* Go 1.26.11 and 1.27.4 are in this Mac's module cache; nothing is downloaded (`GOPROXY=off`), except the one `go install` of golangci-lint v2.15.0 noted in the MADR.
* Heavy runs go through `~/.agents/skills/available-dev-hosts/scripts/run-heavy -- <command>`, one at a time.
* Edits are made by a Python script that asserts each old string's count before replacing, and each result is read back.

## Phase 1 - Records and the change

Files and lines at HEAD:

| File | Line | Old | New |
| :--- | :--- | :--- | :--- |
| `go.mod` | 3 | `go 1.27.4` | `go 1.26.11`, then a blank line and `toolchain go1.27.4` |
| `Makefile` | 1 | `MOD_VERSION := 1.27.4` | `MOD_VERSION := 1.26.11` |
| `docs/README.md` | 29 (the row "know why the module requires Go 1.27.4") | names a current requirement | reworded to point to this pair for the current pin and to 0002 as history |
| `docs/README.md` | table | no row | two index rows for 0007 and two "I want to..." rows |
| `docs/decisions/0002-MADR-pin-go-1.27.4.md` | end | none | `### Amendment 2026-11-17 (second)` pointing here |
| `docs/decisions/0002-PLAN-pin-go-1.27.4.md` | end | none | dated note pointing here |

History that stays: every other line of records 0001 to 0006, including the titles in the index, `go 1.27.4` statements inside records 0002, 0003 and 0005, and the gitignored `dist/` binaries.

Checks, after the edit:

1. `go mod tidy -diff` exits 0, and `git diff -- go.sum` is empty.
2. `go list -m all` shows `mcp-buntdb v1.0.1` and `mcplib v1.6.0`.

## Phase 2 - The version-agreement check

A script kept in the scratch folder, not in the tree, reads `go.mod` for the floor F and the toolchain T and asserts:

* `go.mod` has exactly one `go F` line and one `toolchain goT` line.
* `Makefile:1` is `MOD_VERSION := F`.
* `ci.yml` has no `go-version:` line (both jobs follow `go.mod`) and its golangci-lint install is `@v2.15.0`.
* A sweep: every `1.26.y`, `1.27.y`, `go1.26.y` or `go1.27.y` in a current statement outside `docs/decisions/` equals F or T; and no document outside `docs/decisions/` states `go 1.27.4` as the floor.

It must exit 0 on the changed tree and non-zero on three planted copies made from `git archive HEAD` plus this diff: a wrong toolchain (`toolchain go1.27.3`), a wrong floor (`go 1.26.10`), and a stray `go 1.27.4` floor statement in a doc (here, a line added to `docs/README.md`). A lint pin below v2.15.0 is also planted and must fail.

## Phase 3 - Gates

Each with start, end and exit status captured before any filter, after the last change to a non-record file:

1. `make fmt`, `make vet`, and `make lint` with `GOLANGCI_LINT` set to the scratch v2.15.0 binary, under go1.27.4 and under `GOTOOLCHAIN=go1.26.11`.
2. `make test` on the session toolchain, through `run-heavy`.
3. The floor: `GOTOOLCHAIN=go1.26.11 GOPROXY=off go build ./... && go vet ./...`, then `go test -count=1 -timeout 20m ./...` through `run-heavy`.
4. `govulncheck ./...` with `GOPROXY=off` under go1.27.4 and under go1.26.11; exit 0 is the confirmation.
5. Cross-builds as the Makefile ships them (`GOOS=linux CGO_ENABLED=0`, `GOOS=windows`), into a scratch directory.
6. After the last record edit: `make check-records`, `make markdownlint`, the global disclosure guard's scan for identifiers, and a link check by `check_records.py`.

## Phase 4 - Staging

`git add` only the record files and `docs/README.md` (the index). Every other changed file stays unstaged and is listed to the coordinator. No commit, push or tag.

## Rollback

Before any tag, revert the commits with new commits: restore `go 1.27.4` at `go.mod:3`, remove the `toolchain` line, restore `Makefile:1`, and drop the notes and rows. No history is rewritten. A binary already installed is not changed by this plan.

## Completion Criteria

* `go.mod` says `go 1.26.11` and `toolchain go1.27.4`; `Makefile:1` says `1.26.11`.
* `go.sum` is unchanged and `go mod tidy -diff` exits 0.
* The version-agreement script passes on the tree and fails on all planted copies.
* Every gate in Phase 3 exits 0.
* Staging is as Phase 4 says, and nothing is committed.
* Frontmatter `status` is `implemented` once the coordinator's push and CI pass; until then `in-progress`.

## Execution record

### 2026-11-17 - Phase 1 to Phase 3

Start state: `main` at `7acaf33`, clean, level with `origin/main` (no fetch); `scripts/check_records.py --next` printed 0007.

Phase 1: the edits were made by one Python script that asserted each old string's count. `go.mod` now reads `go 1.26.11`, a blank line, `toolchain go1.27.4`. `Makefile:1` reads `MOD_VERSION := 1.26.11`. `go mod tidy -diff` exited 0 and `go.sum` is unchanged.

Phase 2: the script passed on the tree (`floor=1.26.11 toolchain=1.27.4`). It exited 1 on each planted copy: a wrong toolchain (`go.mod` `toolchain go1.27.3`, naming the 1.27.4 rows of `docs/README.md`), a wrong floor (`go 1.26.10`, naming `Makefile:1` and the 1.26.11 rows of `docs/README.md`), a stray floor statement (`docs/README.md:42: states 'go 1.27.4' as a floor`), and a lint pin of v2.13.1 (`ci.yml:20`).

Phase 3, exit statuses taken before any filter:

| Gate | Toolchain | Exit |
| :--- | :--- | :--- |
| `make fmt` (no file changed) | session | 0 |
| `make vet` | go1.27.4 | 0 |
| `make lint`, golangci-lint v2.15.0 built into a scratch directory (`0 issues.`) | go1.27.4 | 0 |
| `make lint`, same binary (`0 issues.`) | go1.26.11 | 0 |
| `go build ./...` and `go vet ./...`, `GOPROXY=off` | go1.26.11 | 0 and 0 |
| `govulncheck ./...`, `GOPROXY=off` ("No vulnerabilities found.") | go1.27.4 | 0 |
| `govulncheck ./...`, `GOPROXY=off` ("No vulnerabilities found.") | go1.26.11 | 0 |
| `make test` through `run-heavy`, 11 packages `ok` | session (go1.27.4) | 0 |
| `go test -count=1 -timeout 20m ./...` through `run-heavy`, 11 packages `ok` | go1.26.11 | 0 |
| cross-builds linux (`CGO_ENABLED=0`) and windows | go1.27.4 and go1.26.11 | 0 on all four |
| `make check-records`, `make markdownlint` (0 issues) | n/a | 0 and 0 |

The same lint run on a scratch copy of HEAD under go1.27.4 also gave `0 issues.`, so no finding is new with v2.15.0 and none appears only at the floor. golangci-lint v2.15.0 was not in the module cache, so the module proxy served that one download. The two tests that ran through `run-heavy` waited about 34 and 14 minutes in the queue and ran for seconds.

No deviation. Nothing is committed.
