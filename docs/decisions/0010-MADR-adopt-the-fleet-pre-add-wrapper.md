---
status: accepted
date: 2026-11-20
---
# Adopt the fleet pre-add wrapper: golangci-lint over staged packages

## Context and Problem Statement

The fleet's Go pre-add rule — dotfiles
`docs/decisions/0028-MADR-retire-per-file-golint-from-the-fleet-pre-add-rule.md`
(the fleet standard lives in that repository; cited by name, a relative link
cannot reach across repositories) — makes `gofmt` and golangci-lint over the
packages of the staged files the pre-add lint standard, carried by each
repository's pre-add wrapper. This repository's `AGENTS.md` Pre-add section
says "Before staging, run the repository's pre-add command if one exists" —
but no wrapper exists: the scaffold
([0006-MADR-scaffold-agent-workspace.md](0006-MADR-scaffold-agent-workspace.md))
wrote the generic deferral because no script shipped. The machine-wide agent
gate (`~/.agents/hooks/lib/precommit-checks.sh`) prefers a repository's own
`scripts/go-precheck.sh` when one ships, so this repository has no agent
pre-add gate at all, and the fallback the fleet rule names — per-file
`golint` — is an archived tool this repository does not pin.

`golangci-lint` itself is already adopted and pinned:
[0005-MADR-pin-golangci-lint-v2.15.0.md](0005-MADR-pin-golangci-lint-v2.15.0.md)
pins v2.15.0 for `make lint` and CI; self-update and release flows are
unaffected by this record.

## Decision Drivers

* The fleet rule's pre-add standard is golangci-lint-on-staged; every sibling
  repository's wrapper documents that pattern and go-selfupdate-lib's wrapper
  is the recorded port source.
* `golint` upstream is archived; `revive` in this repository's `.golangci.yml`
  is the maintained home of its check family, and this repository never
  invoked golint (dotfiles 0028 finds none here), so adopting the wrapper
  loses nothing.
* The machine-wide agent gate prefers `scripts/go-precheck.sh` when it ships,
  so a repo-local script engages the gate with no hook change.

## Considered Options

* **Adopt the fleet pre-add wrapper** — `scripts/go-precheck.sh` adapted from
  go-selfupdate-lib's, wired as `make pre-add-check` and named in
  `AGENTS.md`. Good, because the pre-add lint becomes the same golangci-lint
  command and config `make lint` runs, so a commit cannot pass a weaker check
  than CI applies; `revive`'s rules already cover the style checks, so
  nothing archived is referenced; the machine-wide agent gate engages
  automatically with no hook change; and staged-file diagnostics keep the
  named-output property the old per-file rule existed for.
  * Bad, because the host-GOOS lint is not stricter than CI's (see the
    decided point below) — but that is today's `make lint` behavior too, so
    the gate introduces no drift.
* **Leave the placeholder deferral** — Good, because nothing changes. Bad,
  because the only tool the fallback names is archived and unpinned here,
  and this repository stays the one fleet member with no pre-add gate; the
  finding dotfiles `0028-MADR` recorded ("under-gated, not redundant") is
  left open.
* **Centralize one script in another repository and link it** — Good,
  because every repository then shares one implementation. Bad, because the
  machine-wide agent gate prefers a repo-local `scripts/go-precheck.sh`, a
  linked script cannot be pinned and audited inside this repository's
  records, and sibling repositories all own their wrappers.

## Decision Outcome

Chosen option: "Adopt the fleet pre-add wrapper" — implement
`scripts/go-precheck.sh` (adapted from go-selfupdate-lib's), a
`make pre-add-check` target, and the `AGENTS.md` Pre-add section, per
[0010-PLAN-adopt-the-fleet-pre-add-wrapper.md](0010-PLAN-adopt-the-fleet-pre-add-wrapper.md).

The lint step runs **once for the host GOOS**: the same command, config file
and GOOS that `make lint` runs, so pre-add and CI cannot drift. A per-GOOS
loop (as go-selfupdate-lib justifies in its
`0002-MADR-rehome-selfupdate-from-mcplib.md` §5) would make pre-add stricter
than this repository's `make lint`, which is host-only (its `Makefile` `lint`
target); making `make lint` per-GOOS is a separate decision this record does
not make.

`govulncheck` runs through the same pinned invocation as the `vuln` target
(`go run golang.org/x/vuln/cmd/govulncheck@v1.10.0 ./...`), so no second
govulncheck version is installed; `GO_PRECHECK_SKIP_VULN=1` skips it offline.
`golint` is retired here: the machine-wide rule that required a per-file
golint pass was retired fleet-wide by the dotfiles record above.

## Consequences

* Good, because this repository stops being the fleet member with no pre-add
  gate; the machine-wide agent gate engages with no hook change once the
  script ships.
* Good, because the pre-add lint is the same command and config as
  `make lint` and CI, so checks cannot drift; golangci-lint stays at the
  [0005-MADR](0005-MADR-pin-golangci-lint-v2.15.0.md) pin.
* Bad, because the host-GOOS choice leaves `*_windows.go` and `*_darwin.go`
  files linted only by CI's single ubuntu job — the same as `make lint`
  today; a per-GOOS tightening is a possible later record.
* Neutral, because `make verify` and CI keep their existing gate lists; this
  record adds a pre-add gate only.

## Confirmation

* `make pre-add-check` planted-finding checks: a deliberately mis-formatted
  staged Go file fails naming the file; a staged file with a revive-reported
  style defect fails naming `file:line`; the clean tree passes 10 cold runs
  (dated evidence in the execution record of
  [0010-PLAN-adopt-the-fleet-pre-add-wrapper.md](0010-PLAN-adopt-the-fleet-pre-add-wrapper.md)).
* `make check-records` and `make markdownlint` pass with the new records and
  docs rows.
* The machine-wide agent gate picks the script up automatically on the next
  agent commit in this repository (it prefers `scripts/go-precheck.sh` when
  it ships), with no hook change — observed once on the first post-change
  agent commit.
