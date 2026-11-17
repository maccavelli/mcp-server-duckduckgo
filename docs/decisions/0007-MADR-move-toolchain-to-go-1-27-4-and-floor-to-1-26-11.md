---
status: accepted
date: 2026-11-17
decision-makers: the repository owner
consulted: "govulncheck v1.9.0 (vulnerability database updated 2026-11-11) under go1.26.11 and go1.27.4, 2026-11-17; go-selfupdate-lib 0045-MADR-move-toolchain-to-go-1-27-4-and-floor-to-1-26-11.md, read-only; actions/setup-go v7.0.0's README as quoted by 0002-MADR-pin-go-1.27.4.md"
informed: the coordinator, for the commit and push through CI
---
# Declare `go 1.26.11` as the floor and `toolchain go1.27.4` as the build toolchain

## Context and Problem Statement

The owner's instruction of 2026-11-17, relayed by the coordinator:

> "go-llmprovider-sdk, go-tui-lib, acp-go-sdk2, gobble-cli, magic-cli-remote, go-selfupdate-lib, mcp-server-duckduckgo, mcp-server-magictools, mcp-server-recall, mcp-server-sequential-thinking, mcplib, mcp-buntdb, and mcp-server-socratic-thinker all need to have their go floor pinned to 1.26.11, and go version/toolchain pinned to 1.27.4."

"Floor" is the `go` directive in `go.mod`. "Version/toolchain" is the `toolchain` line. A read-only survey then found that four of the thirteen need Go 1.27 and cannot build at 1.26.11. The owner's rulings, labels quoted exactly and only as labels:

* "Their floor is 1.27.4: raise go-tui-lib, leave the other three (Recommended)". This repository is not one of the four; it gets floor 1.26.11 and toolchain 1.27.4.
* "One short record pair per repository, all approved now through CI (Recommended)". The pair is committed, executed and pushed to `main` on that approval. The tag and the release wait for the owner's word.
* "v2.15.0 everywhere, in each repository's Go-pins record (Recommended)": each record includes the CI linter pin, and any new finding is brought to the owner rather than suppressed.

This repository is the unusual one of the wave. Its floor today is above the target, so the `go` line goes down, from `go 1.27.4` to `go 1.26.11`, and a `toolchain go1.27.4` line is added. That reverses part of [0002-MADR-pin-go-1.27.4.md](0002-MADR-pin-go-1.27.4.md), which chose `go 1.27.4` with no `toolchain` line.

**Why 0002 chose `go 1.27.4`, and what lowering the floor loses.** Read from 0002's own text, the reasons were: the owner named Go 1.27.4; the previous floor bump (commit `cd66ef28`) was a two-line change to `go.mod` and `MOD_VERSION`; CI reads one version string through `go-version-file: go.mod`; and a `toolchain` line equal to the `go` line is omitted by `go get`, so it was rejected as redundant. It did not choose 1.27.4 for a language or library feature (its Decision Drivers say "nothing in the build list requires a Go newer than 1.26.6"), and it did not mention an advisory. So lowering the floor loses none of its reasons: (a) the build needs nothing from 1.27 (E3, E4); (b) 1.26.11 is free of GO-2026-6859 (E5); (c) the one objection to a `toolchain` line, that it is redundant when equal, no longer holds once the floor and the toolchain differ.

Everything in this record other than the quoted statement and labels is evidence or the author's text.

**Evidence index.** Read or measured on 2026-11-17 at HEAD `7acaf33` on `main`, with a clean tree, level with `origin/main` (no fetch).

* **E1, the pins.** `go.mod:3` is `go 1.27.4`, with no `toolchain` line. `Makefile:1` is `MOD_VERSION := 1.27.4`. A whole-repository search finds no recipe, script or workflow that expands `MOD_VERSION`; it is a documentation-only marker that 0002 kept paired with the `go` line.
* **E2, CI.** `.github/workflows/ci.yml:15` and `:33` read `go-version-file: go.mod` through `actions/setup-go` pinned to commit `b7ad1dad31e06c5925ef5d2fc7ad053ef454303e` (v7.0.0). 0002-MADR quotes that README: a `toolchain` directive, when present, is the version installed; otherwise the `go` directive. With `check-latest` unset (false), the installed release is exactly that version. CI therefore builds both jobs with 1.27.4 after this change, as before. There is no CI job that builds on the floor (a named gap; see Consequences).
* **E3, no Go 1.27 source use.** A search of every `.go` file for `encoding/json/v2` and `jsontext` finds nothing.
* **E4, the floor builds.** On a scratch copy of HEAD with only `go.mod` changed to `go 1.26.11` plus `toolchain go1.27.4`, under `GOTOOLCHAIN=go1.26.11 GOPROXY=off`, `go build ./...`, `go vet ./...` and `go mod tidy -diff` exit 0, and `go.sum` and every `require` line are unchanged. The PLAN records the full-module test run.
* **E5, advisory.** `govulncheck ./...` with `GOPROXY=off` exits 0 ("No vulnerabilities found.") under `GOTOOLCHAIN=go1.26.11` and under `GOTOOLCHAIN=go1.27.4`. GO-2026-6859 (`net/http`) is reported under go1.27.3 and go1.26.10 and under neither of these two, as measured the same day in the sibling repositories.
* **E6, the linter.** `ci.yml:20` already installs `golangci-lint@v2.15.0`, under [0005-MADR-pin-golangci-lint-v2.15.0.md](0005-MADR-pin-golangci-lint-v2.15.0.md). No change is needed. v2.15.0 was built into a scratch `GOBIN` (the module cache lacked it, so the module proxy was used for that one download) and prints `golangci-lint has version 2.15.0 built with go1.27.4`. `make lint` with it exits 0 with `0 issues.` on HEAD under go1.27.4, and on the changed scratch tree under both go1.27.4 and `GOTOOLCHAIN=go1.26.11`. No finding appears at the floor only.
* **E7, rules that state the pin.** `CLAUDE.md` (and `AGENTS.md`, its symlink) states no Go version. `README.md` and `docs/architecture.md` name none. The only current statements are `go.mod`, `Makefile:1`, and `docs/README.md` row "know why the module requires Go 1.27.4".
* **E8, consumers.** No module under the fleet's Go directory requires this module, and its packages live under `internal/` and `cmd/`, so none can. It is a server binary. Its latest tag is `v1.1.0`, whose `go.mod` said `go 1.26.6`.
* **E9, dependencies.** `mcp-buntdb v1.0.1` and `mcplib v1.6.0` are not touched. `go.sum` does not move.

## Decision Drivers

* The owner's instruction and rulings above.
* Ship nothing built on a toolchain with a published advisory: CI must build with 1.27.4, and the declared floor must itself be advisory-free (E5).
* The floor should state the oldest Go the code needs, and the code needs nothing newer than 1.26.11 (E3, E4).
* The tree states one floor and one toolchain, and every current statement agrees with `go.mod`.
* No gate is loosened and no finding is suppressed.
* No compatibility engineering: the owner is the only user of the fleet.

## Considered Options

* **A.** `go 1.26.11` and `toolchain go1.27.4`, as ruled.
* **B.** Keep `go 1.27.4` and add nothing (0002's outcome stands).
* **C.** `go 1.26.11` with no `toolchain` line.
* **D.** `go 1.26.11` and `toolchain go1.27.4` plus a new CI job on the floor.

## Decision Outcome

Chosen option: "A", because it is the owner's instruction applied to a module that builds at 1.26.11 (E4) with an advisory-free floor (E5), and the `toolchain` line keeps CI on 1.27.4 through `go-version-file` (E2). The owner's instruction of 2026-11-17 supersedes the outcome of 0002 for the `go` line and the absence of a `toolchain` line; 0002's other content, its history, and the Makefile marker's pairing with the `go` line stand. This is safe because 0002's reasons (above) are not about Go 1.27, and the one risk 0002 guarded against, an older toolchain loading the module, is now guarded by a floor that is itself advisory-free.

Changes, each with its line at HEAD:

1. `go.mod:3`: `go 1.27.4` becomes `go 1.26.11`, and a blank line and `toolchain go1.27.4` follow it.
2. `Makefile:1`: `MOD_VERSION := 1.27.4` becomes `MOD_VERSION := 1.26.11`, keeping it paired with the `go` line as 0002 and `cd66ef28` did.
3. `docs/README.md`: index rows for this pair, and the "I want to..." row "know why the module requires Go 1.27.4" (a current statement) is reworded so it is true.
4. A dated amendment note in each of 0002's two records pointing here. Nothing in them is rewritten.
5. `ci.yml` is not changed: the lint pin is already v2.15.0 (E6), and both jobs follow `go.mod` (E2).

## Consequences

* Good, because CI keeps building with 1.27.4 and the declared floor is advisory-free.
* Good, because a contributor on Go 1.26.11 can build the module without a toolchain download.
* Good, because `go.sum` and the requirements do not move (E4, E9).
* Neutral, because `MOD_VERSION` is still unused by any recipe.
* Neutral, because a machine whose `GOTOOLCHAIN` names a specific release overrides both lines; the lines then change nothing locally, and CI's setup-go logs show them taking effect.
* Bad, because **CI has no job that builds and tests on the floor.** Both jobs install 1.27.4, so a use of a 1.27-only feature would pass CI while breaking the declared floor. This is a named gap; this change does not add a job. The owner decides whether to add floor jobs fleet-wide. Until then the floor is proved by the PLAN's local run under `GOTOOLCHAIN=go1.26.11` on one machine.
* Neutral, because there is no downstream consumer (E8), so no module's minimum changes.
* Whether a release is needed: **yes, for an installed binary to benefit.** The standard library is compiled into the binary, so a binary built before the toolchain move keeps the standard library it was built with. `v1.1.0` was built when `go.mod` said `go 1.26.6`. A new tag from `main` is the only way an installed binary takes 1.27.4's `net/http`. This record and its PLAN stop before the tag and the release, which each need the owner's word. Nothing here installs, restarts or calls a running server.

## Confirmation

* `go.mod` says `go 1.26.11` and `toolchain go1.27.4`; `Makefile:1` says `1.26.11`; `go.sum` and `go mod tidy -diff` show nothing else moving.
* The PLAN's version-agreement script exits 0 on the tree and non-zero on three planted copies (a wrong toolchain, a wrong floor, a stray `go 1.27.4` floor statement in a doc).
* `govulncheck ./...` exits 0 under go1.26.11 and under go1.27.4, with `GOPROXY=off`.
* `go build ./...`, `go vet ./...` and `go test -count=1 ./...` pass under `GOTOOLCHAIN=go1.26.11`, and `make test`, `make vet` and the lint run pass on the session toolchain.
* The repository's record, link, markdownlint and disclosure checks pass.
* After the push, CI's `validate` and `build` jobs pass and their setup-go logs name 1.27.4.

## Pros and Cons of the Options

### A. Floor 1.26.11, toolchain 1.27.4 (chosen)

* Good, because it is the owner's instruction and CI and the floor are each advisory-free.
* Good, because it declares what the code needs rather than the newest release.
* Bad, because it reverses 0002's `go` line, and CI has no floor job to guard the lowered floor (see Consequences).

### B. Keep `go 1.27.4`, no `toolchain` line

* Good, because nothing changes and CI already builds with 1.27.4.
* Bad, because it disobeys the owner's instruction, and its floor forces a contributor onto 1.27.4 for no feature the code uses.

### C. `go 1.26.11` with no `toolchain` line

* Good, because there is one version line to track.
* Bad, because setup-go then installs the `go` line, so CI would build with 1.26.11 and not 1.27.4, and the owner asked for a 1.27.4 toolchain.

### D. As A plus a CI floor job

* Good, because it would prove the floor on every push.
* Bad, because it is a workflow change outside this change's scope; it is a named gap for the owner to decide fleet-wide, not a part of this record.

## More Information

### Decided

* The target, the one pair per repository, and the linter pin were ruled by the owner on 2026-11-17 as quoted in the context. The reversal of 0002's `go` line is the direct result of the owner's instruction, which names this repository.
* Relationship to earlier records: [0002-MADR-pin-go-1.27.4.md](0002-MADR-pin-go-1.27.4.md) and [0002-PLAN-pin-go-1.27.4.md](0002-PLAN-pin-go-1.27.4.md) get a dated note; [0005-MADR-pin-golangci-lint-v2.15.0.md](0005-MADR-pin-golangci-lint-v2.15.0.md) stands unchanged and is reaffirmed.
* The model is go-selfupdate-lib's `0045-MADR-move-toolchain-to-go-1-27-4-and-floor-to-1-26-11.md`, in a short form. A relative link cannot reach another repository, so it is cited by repository and filename.

### Not verified

* That setup-go v7.0.0 installs the `toolchain` line's version: taken from 0002-MADR's quotation of the README at the pinned commit, not re-read, and no CI run has been seen.
* That CI's own `govulncheck` version, if CI ran one, reports GO-2026-6859: CI here runs no `govulncheck`; only v1.9.0 ran locally.
* That the full test suite passes on a Linux or Windows host at 1.26.11: it runs on one macOS machine only.
* What the other fixes in 1.27.4 and 1.26.11 mean for this module: none was assessed.
