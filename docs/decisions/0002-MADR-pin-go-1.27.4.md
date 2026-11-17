---
status: accepted
date: 2026-11-17
---
# Pin this module to Go 1.27.4

## Context and Problem Statement

The owner asked to pin Go 1.27.4 in this repository and rebuild the current tree.

The module still declares a lower floor. `go.mod` line 3 is `go 1.26.6`, and there is no `toolchain` directive. `Makefile` line 1 is `MOD_VERSION := 1.26.6`. No recipe in that Makefile expands `MOD_VERSION`. Commit `cd66ef28` (`build: raise the Go floor to 1.26.6`) changed those two lines together and left `go.sum` byte-identical.

`.github/workflows/ci.yml` installs Go with `actions/setup-go` `b7ad1dad31e06c5925ef5d2fc7ad053ef454303e` (v7.0.0) and `go-version-file: go.mod` in both the `validate` and `build` jobs. It does not set `go-version` or `check-latest`. The setup-go v6 notes, which v7 kept, say: if a `toolchain` directive is present its version is used; otherwise the action uses the `go` directive. With `check-latest` left at its default `false`, the action installs that exact release.

There is no `go.work`, `.go-version`, `mise.toml`, `.tool-versions`, or Dockerfile. `VERSION` is the product version `3.2.0`. `dist/` is gitignored.

On 2026-11-16 the session host already runs `go version go1.27.4 darwin/arm64`. `go env GOTOOLCHAIN` and `go env GOVERSION` both print `go1.27.4`. That `GOTOOLCHAIN` value is a specific toolchain name, so it overrides the `go` and `toolchain` lines for commands run in this environment ([Go toolchains](https://go.dev/doc/toolchain)).

`go list -m -f '{{.GoVersion}} {{.Path}}' all` on that toolchain shows the highest dependency floor is `1.26.6` (`github.com/maccavelli/mcplib` and this module). The next is `github.com/maccavelli/mcp-buntdb` at `1.26.5`. A main-module `go` line of `1.27.4` is greater than or equal to every required module, which the toolchain spec requires.

The same document states that a `go.mod` saying `go 1.27.4` with no `toolchain` line is read as an implicit `toolchain go1.27.4`. The `go` command uses its bundled toolchain when that toolchain is at least that new, and switches upward only when the line is newer than the bundled toolchain. `go get go@1.27.4` writes the `go` line and omits the `toolchain` line when the two versions are equal.

The working tree also contains the uncommitted Open Library book-search change from [0001-MADR-retarget-book-search-open-library.md](0001-MADR-retarget-book-search-open-library.md). A rebuild compiles that tree. This decision does not reopen 0001.

Which files should record Go 1.27.4, and what should the rebuild compile?

## Decision Drivers

* The owner named Go 1.27.4.
* One version string that CI already reads, matching the last floor bump (`cd66ef28`: `go.mod` and `MOD_VERSION` only).
* The Go toolchain specification: the `go` line is the minimum, an equal `toolchain` line is implied and omitted, and a newer local toolchain is allowed when `GOTOOLCHAIN` is `auto`.
* Dependency floors already measured: nothing in the build list requires a Go newer than 1.26.6, and nothing requires staying on 1.26.6.
* Rebuild the current working tree into `dist/` with the toolchain this host already runs. `dist/` is gitignored.
* Leave the Open Library retarget, product `VERSION`, workflow YAML, and `go.sum` alone unless the version command itself rewrites `go.sum`.

## Considered Options

* Set `go 1.27.4` in go.mod and `MOD_VERSION` to 1.27.4, then rebuild `dist/`
* Add an explicit `toolchain go1.27.4` line beside the `go` line
* Pin 1.27.4 only in the GitHub Actions workflow
* Force `GOTOOLCHAIN=go1.27.4` inside the Makefile recipes

## Decision Outcome

Chosen option: "Set `go 1.27.4` in go.mod and `MOD_VERSION` to 1.27.4, then rebuild `dist/`", because that is the pair commit `cd66ef28` already uses, `go 1.27.4` is the implied toolchain `go1.27.4`, and both CI jobs install the `go` directive exactly while `check-latest` stays false. The session host's `GOTOOLCHAIN=go1.27.4` then builds that tree.

### Consequences

* Good, because CI follows `go.mod` without a second copy of the version in `ci.yml`.
* Good, because `go get go@1.27.4` is the toolchain document's update command, and an equal toolchain line stays omitted.
* Good, because a Go older than 1.27.4 refuses to load the module, or switches up when `GOTOOLCHAIN` allows it.
* Neutral, because a machine whose bundled toolchain is newer than 1.27.4 and whose `GOTOOLCHAIN` is `auto` still compiles with that newer toolchain. The `go` line does not downgrade. CI does not float: setup-go installs 1.27.4.
* Neutral, because `MOD_VERSION` remains unused by recipes. It stays in lockstep with the `go` line so the next reader sees the same pair `cd66ef28` wrote.
* Bad, because `make` invokes `go` on `PATH`. A contributor who has not installed 1.27.4 depends on toolchain download (`GOTOOLCHAIN=auto`) or gets a refusal (`GOTOOLCHAIN=local`).

### Confirmation

`go.mod` contains the line `go 1.27.4` and no `toolchain` line. `Makefile` assigns `MOD_VERSION := 1.27.4`. `go.sum` is unchanged. `go env GOVERSION` in this module prints `go1.27.4` on the session host. `make build-all` succeeds, and `go version -m` on each `dist/` binary reports `go1.27.4`.

## Pros and Cons of the Options

### Set `go 1.27.4` in go.mod and `MOD_VERSION` to 1.27.4, then rebuild `dist/`

The same two-line edit as `cd66ef28`, then `make build-all` so the gitignored `dist/` binaries match the working tree, including the uncommitted Open Library client.

* Good, because CI's `go-version-file: go.mod` picks up the new floor with `check-latest` false, so the installed release is 1.27.4.
* Good, because the implied toolchain name is `go1.27.4`, which matches the session host.
* Good, because every module in `go list -m all` declares `go` 1.26.6 or older, so 1.27.4 is a legal main-module floor.
* Neutral, because `MOD_VERSION` does not affect `go build`.
* Bad, because `go.mod` cannot force a downgrade when a newer toolchain is already the bundled default and `GOTOOLCHAIN=auto`.

### Add an explicit `toolchain go1.27.4` line beside the `go` line

* Good, because setup-go v7 prefers the `toolchain` directive when it is present, and the version would still be 1.27.4.
* Bad, because the toolchain spec says an equal toolchain line is omitted, and `go get go@1.27.4` deletes it.
* Bad, because the line does not pin harder than the `go` line: selection still uses a bundled toolchain that is already newer.

### Pin 1.27.4 only in the GitHub Actions workflow

Set `go-version: '1.27.4'` and leave `go 1.26.6` in `go.mod`.

* Good, because CI would install 1.27.4 even if `go.mod` lagged.
* Bad, because the module would still advertise 1.26.6, so a local Go 1.26.6 would load it while CI built with 1.27.4.
* Bad, because `go-version` overrides `go-version-file` when both are set, which splits the version across two files. The workflow already has a single source.

### Force `GOTOOLCHAIN=go1.27.4` inside the Makefile recipes

* Good, because every `make` build on any host would run the named toolchain, downloading it when `auto` semantics of a bare name apply.
* Bad, because a specific `GOTOOLCHAIN` name overrides `go.mod` and a deliberate newer local toolchain.
* Bad, because `cd66ef28` did not put the toolchain in Make, and the `go` line already makes older toolchains refuse or switch.

## More Information

Toolchain selection and `go get go@1.27.4`: [https://go.dev/doc/toolchain](https://go.dev/doc/toolchain). setup-go v7.0.0 README at the pinned commit: `toolchain` wins when present, otherwise the `go` directive; `check-latest` defaults to false.

Implementation plan: [0002-PLAN-pin-go-1.27.4.md](0002-PLAN-pin-go-1.27.4.md).

### Amendment 2026-11-17

`go get go@1.27.4` also regrouped `require` blocks. Versions, `// indirect` marks, and `go.sum` stayed the same. The owner kept that layout. The decision is still `go 1.27.4` plus `MOD_VERSION := 1.27.4`, with no `toolchain` line.

The book-search work stays as accepted in [0001-MADR-retarget-book-search-open-library.md](0001-MADR-retarget-book-search-open-library.md). This record does not commit either change. HEAD is `main`. A commit needs a later explicit request that names `main`, or a named `feature/…` branch.
