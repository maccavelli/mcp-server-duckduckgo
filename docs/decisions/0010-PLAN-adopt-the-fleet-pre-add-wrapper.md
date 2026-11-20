---
status: "implemented"
date: 2026-11-20
associated-madr: "0010-MADR-adopt-the-fleet-pre-add-wrapper.md"
---
<!-- markdownlint-disable MD013 MD024 -->

# Implement the fleet pre-add wrapper

Associated MADR: [0010-MADR-adopt-the-fleet-pre-add-wrapper.md](0010-MADR-adopt-the-fleet-pre-add-wrapper.md)

## Goal

Ship `scripts/go-precheck.sh` and wire it as `make pre-add-check`, replacing
the `AGENTS.md` deferral ("run the repository's pre-add command if one
exists") with the concrete command, so this repository stops being the fleet
member with no agent pre-add gate. Out of scope: CI workflow changes
(`make verify` already gates CI), `.golangci.yml` changes, per-GOOS
linting (a separate decision), and any `golint` re-introduction.

## Grounding Reference

* [0010-MADR-adopt-the-fleet-pre-add-wrapper.md](0010-MADR-adopt-the-fleet-pre-add-wrapper.md)
  (this decision: host-GOOS lint chosen to match `make lint` exactly; the
  pinned `govulncheck` invocation).
* go-selfupdate-lib's `scripts/go-precheck.sh` — the recorded port source for
  the fleet's wrapper pattern; dotfiles
  `docs/decisions/0028-MADR-retire-per-file-golint-from-the-fleet-pre-add-rule.md`
  — the fleet decision that retires golint and standardizes
  golangci-lint-on-staged.
* [0005-MADR-pin-golangci-lint-v2.15.0.md](0005-MADR-pin-golangci-lint-v2.15.0.md)
  — linter version for both `make lint` and the wrapper's install hint.

## Prerequisites

* `go` (toolchain from `go.mod`), `bash`, `python3`, `npx`, and
  golangci-lint v2.15.0 on `$(go env GOPATH)/bin` or via `GOLANGCI_LINT=`.
* A scratch clone for the planted-finding experiments
  (`/tmp/ddg-scratch-<date>`), never the live tree.

## Phase 1 — records and wrapper script

1. Draft this record pair (the MADR first, this file as its plan).
2. Write `scripts/go-precheck.sh`, ported from go-selfupdate-lib's wrapper:
   * Collect staged Go files (arguments, or every tracked file with none); a
     staged deletion still names its package; a docs-only change exits 0
     before any tool runs.
   * Step 1 `gofmt -l` over the files; gofmt's own failure fails the check.
     Its exit status counts as well as its list (the source script's
     behavior).
   * Step 2 one `golangci-lint run -c .golangci.yml ./...` for the host GOOS —
     the command, config and GOOS `make lint` runs, so the two cannot drift;
     `GOLANGCI_LINT=<path>` overrides the binary; a missing binary fails with
     the `0005-MADR` install line (v2.15.0).
   * Step 3 `go vet` and `go test` over the packages the files belong to,
     with the nearest-`go.mod` module handling the sibling script implements
     (a package with no Go file left widens to `./...`).
   * Step 4 `go run golang.org/x/vuln/cmd/govulncheck@v1.10.0 ./...` — the
     `vuln` target's own pinned invocation, so no second govulncheck
     version; exit 3 (vulnerabilities found) and network-looking failures
     keep the sibling script's distinct exit codes;
     `GO_PRECHECK_SKIP_VULN=1` skips it offline.
3. Wire the repository: add the `pre-add-check` phony target to `Makefile`
   (`.PHONY` line included) invoking `bash scripts/go-precheck.sh $(FILES)`;
   replace the `AGENTS.md` Pre-add deferral with the concrete invocations,
   the step list, the golint retirement note, and the machine-gate
   paragraph; index this record pair in `docs/README.md`.

## Phase 2 — planted-finding experiments and gates

1. Copy the working tree to a scratch clone. On the scratch clone:
   * an unformatted, otherwise trivial staged Go file:
     `bash scripts/go-precheck.sh <file>` exits non-zero naming the file;
   * a staged file with a revive-reported style defect (a superfluous
     `else` after a `return`, which `indent-error-flow` reports) exits
     non-zero naming `file:line`;
   * revert the plants, then ten cold runs exit 0.
2. On the live tree: `make check-records`, `make markdownlint`,
   `make pre-add-check` green; `git status` shows only this record pair, the
   script, the `Makefile`, `AGENTS.md` and `docs/README.md` changed.

## Rollback

Revert this record pair's commit. The tree returns to the scaffold-default
`AGENTS.md` deferral with no `scripts/go-precheck.sh` and no
`make pre-add-check`; the machine-wide agent gate falls back to its own
behavior (the wrapper was an addition, and nothing else references it). No
data migration, no binary changes, no CI edits to unwind.

## Completion Criteria

1. `make check-records`, `make markdownlint`, `make pre-add-check` all pass
   on the changed tree with only the intended files changed.
2. The planted-finding fences above observed red on the scratch clone, then
   clean, 10 of 10 cold runs.
3. `make verify` still green (the wrapper touches nothing it runs).

## Execution record

Populated as phases run; every deviation gets a dated entry here.
