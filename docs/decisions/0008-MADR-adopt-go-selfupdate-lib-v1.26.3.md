---
status: proposed
date: 2026-11-17
consulted: "go-selfupdate-lib tag v1.26.3, peeled commit 1dffdc9dee24a7534513dc51c9039408be37c09b, read 2026-11-17; a scratch module outside this repository built and tested against that tag the same day"
---
# Adopt go-selfupdate-lib v1.26.3 inside the existing update command

## Context and Problem Statement

`mcp-server-duckduckgo update` still constructs its updater from `github.com/maccavelli/mcplib/selfupdate`. The release job still calls mcplib's reusable publish workflow at a pin commented as mcplib v1.4.1, and it still passes `bridge-release: false`. The maintained library and the maintained publish workflow are `github.com/maccavelli/go-selfupdate-lib` at tag `v1.26.3`. [0003-MADR-update-dependency-pins.md](0003-MADR-update-dependency-pins.md) left that replacement as a separate decision. This record is that decision.

The command has to keep working as a Cobra subcommand of an MCP server. `update` must skip configuration, write no JSON-RPC on the protocol stdout, return errors to `main` so the in-process tests survive, and install a standalone binary for the three platforms this repository already publishes.

**Evidence index.** Read or measured on 2026-11-17. This repository was at `035e171de3cd215bde2b4c782e3a31d9de643a0c` on `main`, with a clean tree. The local `origin/main` ref was that same commit. Nothing was fetched in this pass.

* **E1, this module's floor.** `go.mod` says `go 1.26.11` and `toolchain go1.27.4`. That is the outcome of [0007-MADR-move-toolchain-to-go-1-27-4-and-floor-to-1-26-11.md](0007-MADR-move-toolchain-to-go-1-27-4-and-floor-to-1-26-11.md). Direct requirements include `mcplib v1.6.0`, `cobra v1.10.2`, `playwright-go v0.6401.0`, and `golang.org/x/sys v0.50.0`. Indirect requirements include `golang.org/x/mod v0.43.0`, `golang.org/x/term v0.48.0`, and `github.com/ysmood/fetchup v0.3.0`.
* **E2, where mcplib is imported.** `github.com/maccavelli/mcplib/selfupdate` is imported by `cmd/mcp-server-duckduckgo/update.go`, `update_test.go`, and `main.go`. The parent module `github.com/maccavelli/mcplib` is imported by `serve.go`, `serve_test.go`, `internal/util/mcp_helpers.go`, `internal/util/mcp_helpers_test.go`, `internal/handler/system/tools.go`, `internal/handler/system/tools_test.go`, `internal/handler/media/tools.go`, and `internal/handler/search/tools.go`. Those call sites are the process harness: server construction, the stdio pipeline, logging, diagnostic tools, and shutdown classification.
* **E3, the command today.** `update.go` annotates the command `selfupdate.skip-config=true`. `root.go` `PersistentPreRunE` returns before `loadConfig` when that annotation is `true`. `Execute` saves `os.Stdout` in `OriginalStdout`, then points `os.Stdout` at `os.Stderr`, and points the root command's output and error streams at `os.Stderr`. `Execute` returns the error. `main` maps it with `selfupdate.ExitCode` and prints it to stderr when the code is anything other than 10, then calls `os.Exit`. `RunE` does not call `os.Exit`. The tests call `ExecuteContext` in-process.
* **E4, the updater today.** `defaultUpdater` builds `NewGitHubSource` for `maccavelli/mcp-server-duckduckgo`, with an `http.Client` whose `Timeout` is 15 minutes, a user agent of `mcp-server-duckduckgo/` plus `RawVersion`, `DefaultLimits`, `NewExactAssetSelector` of linux/amd64, darwin/arm64, and windows/amd64, `NewStandaloneInstaller`, `NewStrictVersionPolicy`, `NewTextReporter` on the command error stream, and `NewTerminalConfirmer` on `os.Stdin`. `RunE` then applies its own 15-minute `context.WithTimeout` and calls `Updater.Run`. Flags are `--check`, `--yes`/`-y`, `--force`, and `--version`.
* **E5, identity today.** `version.go` defines `RawVersion` default `"dev"`, `RawBuildKind` default `"local"`, and `Version` as `RawVersion` with a leading `v` removed. `buildKind` returns `ReleaseBuild` only when `RawBuildKind` is exactly `release`. The four Makefile recipes pass `-X main.RawVersion=$(VERSION)` and `-X main.RawBuildKind=$(BUILD_KIND)`. `BUILD_KIND` defaults to `local`. The tag job in `.github/workflows/ci.yml` runs `make build-all VERSION=<ref name> BUILD_KIND=release`. `root.go` sets `RootCmd.Version` to the trimmed `Version`. `serve.go` logs that string and passes it to `mcplib.NewMCPServer`.
* **E6, tests that lock the command.** `update_test.go` locks the skip-config annotation, an empty home after `--check`, configuration still loading for any other command, the updater writer being the command error stream, the four flag names and `-y`, rejection of a positional argument, `--check` with `--yes` or `--force` returning an error that is not `ErrUpdateAvailable`, a cancelled caller context, the `release` / `local` / empty / `Release` kind map, the default stamp failing `NewStrictVersionPolicy`, exit codes 0 / 10 / 1, and the three platforms.
* **E7, the publish job today.** `.github/workflows/ci.yml` job `release` calls `maccavelli/mcplib/.github/workflows/publish-selfupdate-release.yml@8d4d89ff34ff79c1ed65f223cc3d83b3ec0550aa`, with `artifact-name`, `products-json` of `mcp-server-duckduckgo`, the three platforms, `extra-assets-json` of `[]`, and `bridge-release: false`. Job permissions are `contents: write`, `id-token: write`, and `attestations: write`. The job runs when `github.ref` starts with `refs/tags/v`.
* **E8, the library tag.** Tag `v1.26.3` is annotated object `10316bca82c00aeefde4bc614d183d343ed5818e`. The peeled commit is `1dffdc9dee24a7534513dc51c9039408be37c09b`, message `docs(release): prepare v1.26.3 changelog and migration guide`, committer date 2026-11-17 14:01:01 -0600. That commit's `go.mod` says `go 1.26.11`, `toolchain go1.27.4`, and requires `golang.org/x/mod v0.42.0`, `golang.org/x/sys v0.49.0`, and `golang.org/x/term v0.47.0`. It retracts `[v1.0.0, v1.4.1]`. The checkout that contained the tag was two documentation commits ahead of the peel. Those commits do not change `go.mod`. This decision uses the peeled commit.
* **E9, the publish workflow at that peel.** `.github/workflows/publish-selfupdate-release.yml` declares `artifact-name`, `products-json`, `platforms-json`, and `extra-assets-json` as required, and `prerelease-channels-json` as optional with default `[]`. Its header comment says `[]` publishes stable tags only. `sign` defaults to false. A search of that file for `bridge-release` found no match. The workflow's own permissions are `contents: write`, `id-token: write`, and `attestations: write`.
* **E10, symbols read at the peel, then executed in a scratch module.** The scratch module lived outside this repository. Its `go` line was `1.26.11` and its toolchain line was `go1.27.4`. `go test -count=1` exited 0. Measured results:
  * `buildinfo.VersionVar` is `github.com/maccavelli/go-selfupdate-lib/buildinfo.version`. `buildinfo.KindVar` is `github.com/maccavelli/go-selfupdate-lib/buildinfo.kind`. `buildinfo.LDFlags("v1.2.3")` is `-X github.com/maccavelli/go-selfupdate-lib/buildinfo.version=v1.2.3 -X github.com/maccavelli/go-selfupdate-lib/buildinfo.kind=release`.
  * `cli.DefaultTimeout` is 15 minutes. `cli.Run`'s comment at the peel says it cancels on the signals and the timeout and never calls `os.Exit`. A scratch assignment of `cli.Run` to `func(context.Context, *selfupdate.Updater, selfupdate.Request, cli.Options) (selfupdate.Result, error)` compiled.
  * `DefaultHTTPClient().Timeout` is 0. The function body at the peel sets `ResponseHeaderTimeout` to 30 seconds. `UserAgent("mcp-server-duckduckgo", "v1.2.3")` returned `mcp-server-duckduckgo/v1.2.3 (` plus the scratch process's GOOS, `/`, and GOARCH, plus `)`.
  * `NewGitHubSource` accepted `DefaultHTTPClient`, that user agent, and `DefaultLimits`, and did not dial. `New` accepted that source, `NewStrictVersionPolicy`, `NewExactAssetSelector` of one platform, `NewStandaloneInstaller`, `DiscardReporter`, and `NonInteractiveConfirmer`.
  * `NewStrictVersionPolicy` rejected `dev` and `v1.2.3-dirty`, and accepted `v1.2.3`.
  * `cli.Flags.Bind` on a Cobra command registered `--check`, `--yes`, `--force`, `--dry-run`, `--json`, `--version`, `--channel`, and shorthand `-y` on `--yes`.
  * `Flags.Request` mapped `buildinfo.KindRelease` to `selfupdate.ReleaseBuild` and `buildinfo.KindLocal` to `selfupdate.LocalBuild`, and set `CurrentVersion` from `Info.Current()`. `--check` with `--yes`, `--force`, or `--dry-run` returned `selfupdate: --check and --yes are contradictory`, `selfupdate: --check and --force are contradictory`, and `selfupdate: --check and --dry-run are contradictory`. None of those errors is `ErrUpdateAvailable`.
  * `cli.Summary` of an available update was `mcp-server-duckduckgo: update available: v1.0.0 -> v1.1.0`. `cli.Summary` of a generic error was empty. `cli.Exit` wrote nothing and returned 10 for `ErrUpdateAvailable`, and wrote `update failed: boom\n` and returned 1 for a generic error.
  * A Cobra command with `SilenceUsage` and `SilenceErrors` set, and the same flags bound, accepted `--help` with a nil error. The help text contained `--check` and `--dry-run`.
  * `selfupdate.ExitCode` returned 0, 10, and 1 for a nil error, `ErrUpdateAvailable`, and a generic error.
* **E11, stamp rule, executed.** A scratch `main` that prints `buildinfo.Identity()` was built with linker flags. Output, one line each:
  * no flags: `version=` empty, `kind=local`, `current=dev`, `reason=` empty.
  * version `v9.9.9`, kind `release`: `kind=release`, `current=v9.9.9`.
  * version `v9.9.9-dirty`, kind `release`: `kind=local`, reason `buildinfo: stamped version "v9.9.9-dirty" is not a release tag`.
  * version `v9.9.9`, kind `Release`: `kind=local`.
  * version `v9.9.9`, kind `local`: `kind=local`.
  * version `v9.9.9-rc.1`, kind `release`: `kind=release`.
  * version empty, kind `release`: `kind=local`, `current=dev`, reason `buildinfo: stamped version "" is not a release tag`.
  * The release expression in `buildinfo` at the peel is `^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[a-z][a-z0-9]{0,15}\.(0|[1-9][0-9]*))?$`, and the kind string that can match is exactly `release`. `Current()` is the stamped version, otherwise the module version unless it is `(devel)`, otherwise `dev`.
* **E12, minimal versions, executed.** After `go mod tidy` on a scratch module whose only requirement was the library, `golang.org/x/sys` was `v0.49.0`. After that module also required `golang.org/x/sys v0.50.0` (imported), `golang.org/x/mod v0.43.0`, and `golang.org/x/term v0.48.0`, `go list -m` showed those three versions and `github.com/maccavelli/go-selfupdate-lib v1.26.3`. The `go` and `toolchain` lines stayed `1.26.11` and `go1.27.4`. `go test -count=1` exited 0.
* **E13, the migration guide at the same peel.** `docs/guides/migrating-from-mcplib-selfupdate.md` section "From v1.26.2 to v1.26.3" says `v1.26.3` changes no exported API, that `make apicheck` reports compatibility with `v1.26.2`, that the floor is Go 1.26.11 because of GO-2026-6859 in `net/http` reached through `GitHubSource`, that the toolchain is Go 1.27.4, and that callers move workflow pins to `v1.26.3`. The same file's closing check still says `go list -m` gives `v1.26.2`. Section 5's sample `go get` names `@v1.26.2`, and its sample workflow pin is `e62d77cddaf264cc6ec7bbc4f39b90322542f752`. `git rev-parse v1.26.2^{}` is that commit. Section 5's canonical command is `cli.Command` for `prepare-commit-msg`, with six platforms. The "With cobra" sample binds `cli.Flags`, calls `cli.Run` and `cli.Exit`, and calls `os.Exit` inside `RunE`. It also says an MCP server that points `os.Stdout` at stderr passes the real stdout in `Options.Stdout` when `--json` is set. `make apicheck` was not run here. The guide's later changelog entries were not re-run here.
* **E14, local-build check, read at the peel.** `classifyOperation` returns `OperationReplaceLocal` for `CheckOnly` on `LocalBuild`. `TestRunLocalBuildCheckHintsForce` expects `ErrUpdateAvailable` and a selected-event detail that mentions `--force`. An apply of a local build without `--force` returns `ErrForceRequired`. This repository's test binary is unstamped, so a `--check` under `go test` is a local-build check.
* **E15, what this record did not run.** No live GitHub update. No tag publish. No `make apicheck`. No run of the library's own test suite. The scratch module did not import `selfupdate/service`.

## Decision Drivers

* Use the current published library for the flag set, the contradiction checks, the run, the build identity, the HTTP client, and the publish workflow.
* Keep the Cobra command, the `selfupdate.skip-config` annotation, the protocol-stdout trap, `main`'s single exit mapping, the in-process tests, `NewStandaloneInstaller`, and the three published platforms.
* Keep `mcplib v1.6.0` for the process harness. Self-update is not its last import (E2).
* Leave the `go` and `toolchain` lines where 0007 put them. The library tag declares the same pair (E8, E12).
* A `uses:` pin is a commit SHA. The SHA is the peeled `v1.26.3` commit, because the annotated tag object is not that commit (E8).
* The first proof of the publisher is a future tag. This decision does not create a tag.

## Considered Options

* **A. Library surface inside the existing Cobra command.** `cli.Flags.Bind`, `Flags.Request`, `cli.Run`, `buildinfo` stamps, `DefaultHTTPClient`, `UserAgent`, `DiscardReporter` and `NonInteractiveConfirmer` as the updater's placeholders, `NewStandaloneInstaller`, the three platforms, and the peeled `v1.26.3` workflow. `main` keeps `selfupdate.ExitCode` and one stderr print.
* **B. Import-path swap only.** Point the current constructors at `go-selfupdate-lib`, and keep the hand-rolled flags, the 15-minute `http.Client` timeout, the hand-rolled context timeout, `NewTextReporter`, `NewTerminalConfirmer`, `RawVersion` / `RawBuildKind`, and the mcplib workflow pin.
* **C. Replace the command with `cli.Command`.** The guide's primary sample: one call parses flags, runs the update, prints the failure line, and returns the exit code.
* **D. Stay on `mcplib/selfupdate`.** Leave `update.go`, `version.go`, the Makefile stamps, and the mcplib workflow pin as they are.

## Decision Outcome

Chosen option: "A. Library surface inside the existing Cobra command.", because that option uses the library for the flags, the contradiction checks, the run, the stamps, the HTTP client, and the publisher, and it leaves the Cobra command, the skip-config annotation, the stdout trap, the in-process tests, and the mcplib server harness in place.

The concrete shape:

1. Require `github.com/maccavelli/go-selfupdate-lib v1.26.3`. `go 1.26.11`, `toolchain go1.27.4`, `mcplib v1.6.0`, `x/sys v0.50.0`, `x/mod v0.43.0`, and `x/term v0.48.0` stay (E1, E12). `fetchup` stays `v0.3.0`. Playwright stays on `github.com/mxschmitt/playwright-go`.
2. `update.go` imports `go-selfupdate-lib/selfupdate`, `go-selfupdate-lib/selfupdate/cli`, and `go-selfupdate-lib/buildinfo`. The command stays Cobra, with `cobra.NoArgs`, the skip-config annotation, `SilenceUsage: true`, and `SilenceErrors: true`. `cli.Flags.Bind` registers the flag set (E10). `RunE` calls `Flags.Request` with `buildinfo.Identity()` before it constructs an updater. `Request`'s error is returned as-is.
3. The updater constructor takes no writer. It matches the guide's `defaultNewUpdateUpdater` with this repository and these three platforms: `NewGitHubSource` using `DefaultHTTPClient()`, `UserAgent(name, Identity().Current())`, `DefaultLimits`, `NewExactAssetSelector` of the frozen three, `NewStandaloneInstaller` of an empty `InstallOptions`, and `New` with `NewStrictVersionPolicy`, `DiscardReporter`, and `NonInteractiveConfirmer`. `cli.Run` installs the real reporter and confirmer through `WithReporter` and `WithConfirmer`.
4. `RunE` calls `cli.Run` with the command context, substituting `context.Background()` when the command context is nil. Timeout stays 0 so `cli.Run` applies `DefaultTimeout`. Signals stay nil so `cli.Run` uses interrupt and `SIGTERM`. `HandOff` stays the zero value. `RunE` returns `cli.Run`'s error. It does not call `cli.Exit` or `os.Exit`.
5. Options: stderr is the command error stream. Stdin is the command stdin. `Interactive` is `cli.StdioOptions().Interactive` when that stdin is `os.Stdin`, and false otherwise. `Options.Stdout` stays nil unless `--json` is set. On `--json` it is a new `protocolStdout` writer, which `Execute` sets to `OriginalStdout` before the stdout alias. A nil `protocolStdout` with `--json` returns an error before the updater is constructed. `cli.StdioOptions().Stdout` is not used: after `Execute`, `os.Stdout` is the aliased stderr (E3, E13).
6. `main.go` changes its selfupdate import to `go-selfupdate-lib/selfupdate` and otherwise keeps the exit mapping in E3. `cli.Exit`'s `update failed:` line is not added. `cli.Run` already writes `Summary` for an available update, a decline, a dry run, and an up-to-date result, and `Summary` is empty for every other error (E10). With `SilenceErrors` set, Cobra does not print a second copy. A failed run can still write an `EventFailed` line through the text reporter before `main` prints the error. That event line is the library's progress line. `--help` still prints, as measured with both silence fields set (E10).
7. Delete `version.go`. The Makefile's four recipes replace `-X main.RawVersion` and `-X main.RawBuildKind` with `-X github.com/maccavelli/go-selfupdate-lib/buildinfo.version=$(VERSION)` and `-X github.com/maccavelli/go-selfupdate-lib/buildinfo.kind=$(BUILD_KIND)`. `Execute` sets `RootCmd.Version` to `Identity().Current()` with one leading `v` removed, which is the display string `serve.go` already passes to the log and to `NewMCPServer` (E5). The update request uses `Identity().Current()` with the `v` kept, because that is what `Flags.Request` puts in `CurrentVersion` (E10). A release identity requires both the kind stamp `release` and a version that matches the release expression (E11). CI's tag build already passes the tag name and `BUILD_KIND=release` (E5).
8. The release job `uses:` line becomes `maccavelli/go-selfupdate-lib/.github/workflows/publish-selfupdate-release.yml@1dffdc9dee24a7534513dc51c9039408be37c09b` with a trailing comment `v1.26.3`. `bridge-release` is removed. `artifact-name`, `products-json`, `platforms-json`, and `extra-assets-json` stay. `prerelease-channels-json` and `sign` are left unset, so the workflow defaults apply (E9). Permissions stay. Checkout, setup-go, upload-artifact, and the golangci-lint pin stay.
9. Tests keep the skip-config, empty-home, other-command, positional, exit-code, and platform locks, retargeted at `go-selfupdate-lib`. The kind and version tests follow E11. The flag test includes `--dry-run`, `--json`, and `--channel`. Contradiction cases include `--check --dry-run` and assert that the updater constructor was not called. One fake-source test, built with `selfupdate/selfupdatetest` and limited to `--check` and `--check --json`, locks the stream split. No test applies an update onto the test binary.

Out of this decision: `selfupdate/service` and any `HandOff` hook; `cli.Command`; the guide's six-platform list; a prerelease channel; signed checksum notes; a tag or a GitHub release; Playwright, fetchup, stealth, and book search.

### Confirmation

Compliance is the completion criteria of [0008-PLAN-adopt-go-selfupdate-lib-v1.26.3.md](0008-PLAN-adopt-go-selfupdate-lib-v1.26.3.md). The API facts in E10, E11, and E12 are already measured. The command tests, `gofmt`, `make lint`, and the module test run happen when that plan is executed. The publish workflow is confirmed by the next tag the owner asks for, which is outside this plan.

## Consequences

* Good, because the flag set, the three contradiction errors, the 15-minute run bound, signal cancellation, and the release-stamp rule come from the library that the scratch module compiled (E10, E11).
* Good, because `Request` refuses a contradictory flag set before an updater exists, which is earlier than today's `Updater.Run` check (E4, E6, E10).
* Good, because `DefaultHTTPClient` has no `Client.Timeout`, so the run context is the bound on a download. The current constructor sets a 15-minute client timeout and a 15-minute context together (E4, E10).
* Good, because `mcplib v1.6.0` stays for the harness, and the floor and toolchain stay on the 0007 pair (E1, E2, E12).
* Good, because `--json` can write to the stdout `Execute` saved, while every other update byte stays on the command error stream (E3, E13).
* Good, because `--help` still prints when `SilenceUsage` and `SilenceErrors` are set (E10).
* Neutral, because a failure line from `main` is the error text. `cli.Exit` would prefix `update failed:`, and calling it from `RunE` as well as `main` would print twice. `SilenceErrors` is what keeps Cobra from adding a third copy.
* Neutral, because a local-build `--check` returns `ErrUpdateAvailable` and the library's selected event mentions `--force` (E14). Exit code 10 still means an actionable check. Applying that local build still takes `--force`.
* Neutral, because `v9.9.9-rc.1` with kind `release` is `KindRelease` (E11). This workflow does not pass `prerelease-channels-json`, so the called workflow's default `[]` stays in force (E9). Publishing a prerelease tag is a later decision.
* Neutral, because an unstamped `go install` can report the module version through `Current()`, while an unstamped `go build` in this module reports `dev` (E11). The trimmed display string follows `Current()`.
* Bad, because a binary stamped `BUILD_KIND=release` whose `VERSION` is not a release tag is a local build (E11). Today's `buildKind` treats any exact `release` stamp as `ReleaseBuild`, including a dirty `git describe`. CI tag builds pass the tag name, so a published tag binary still qualifies. A local `make build BUILD_KIND=release` on a dirty describe does not.
* Bad, because the new publisher cannot be proved before a tag. The plan stops before any tag.
* Bad, because the guide's own closing check still names `v1.26.2`, and section 5 still shows that release's workflow SHA (E13). Following those lines would pin the previous release. The pin in this decision is the peeled `v1.26.3` commit.

## Pros and Cons of the Options

### A. Library surface inside the existing Cobra command

* Good, because the scratch module already built `Flags.Bind`, `Flags.Request`, `cli.Run`, `DefaultHTTPClient`, `UserAgent`, `NewStandaloneInstaller`, and the stamp rule against `v1.26.3` (E10, E11).
* Good, because the skip-config annotation, the saved stdout, and `main`'s exit mapping stay, which is what the current tests and the MCP pipe require (E3, E6).
* Neutral, because the failure line stays `main`'s print of the error text, and the library's `update failed:` prefix stays unused.
* Bad, because the release rule is stricter than the string compare in `buildKind` (E5, E11).

### B. Import-path swap only

* Good, because `NewGitHubSource`, `NewExactAssetSelector`, `NewStandaloneInstaller`, `NewStrictVersionPolicy`, `ExitCode`, and `ErrUpdateAvailable` keep their names, so the current `update.go` compiles after the import edit.
* Neutral, because `go.sum` still changes, and the tests still move to the new module path.
* Bad, because the 15-minute `Client.Timeout`, the hand-rolled flags, and the `main.RawVersion` stamps remain, which is the surface `buildinfo` and `cli` exist to replace (E4, E10).
* Bad, because the publish job would keep the mcplib workflow pin and `bridge-release`, while the workflow at `v1.26.3` has no such input (E7, E9).

### C. Replace the command with `cli.Command`

* Good, because one library call then owns flag parsing, help text, the `update failed:` line, signals, and the timeout (E13).
* Neutral, because the guide demonstrates that call on `prepare-commit-msg`, whose `main` is not this Cobra tree.
* Bad, because `cli.Command` has no way to carry `selfupdate.skip-config`, so configuration would load on `update` unless a Cobra wrapper remained, at which point the wrapper is option A.
* Bad, because the guide's Cobra sample calls `os.Exit` inside `RunE`. The tests in E6 run the command in-process. An `os.Exit` there ends the test process.
* Bad, because `cli.StdioOptions().Stdout` after `Execute` is the aliased stderr, so `--json` would miss the saved protocol stdout (E3, E13).

### D. Stay on `mcplib/selfupdate`

* Good, because the current tests already pass on `main` and no publish pin moves.
* Neutral, because `mcplib v1.6.0` remains a dependency under every option, for the harness (E2).
* Bad, because the updater, the stamps, and the publisher stay on the older surface, including a workflow input the maintained workflow does not declare (E4, E7, E9).
* Bad, because 0003 already recorded this replacement as a separate decision, and the library tag now matches the floor and toolchain this module declares (E1, E8).

## More Information

* The deferred sentence is the last paragraph of [0003-MADR-update-dependency-pins.md](0003-MADR-update-dependency-pins.md). The implementation plan adds a dated pointer there when this decision is accepted. The paragraph itself stays.
* The floor and toolchain this adoption relies on are [0007-MADR-move-toolchain-to-go-1-27-4-and-floor-to-1-26-11.md](0007-MADR-move-toolchain-to-go-1-27-4-and-floor-to-1-26-11.md).
* Library guide, cited by repository and filename because a relative link cannot cross repositories: `go-selfupdate-lib` `docs/guides/migrating-from-mcplib-selfupdate.md` at peeled commit `1dffdc9dee24a7534513dc51c9039408be37c09b`. The workflow file is `.github/workflows/publish-selfupdate-release.yml` at that same commit.
* `make apicheck` is the guide's claim in E13. It is not a measurement of this record.
* Implementation plan: [0008-PLAN-adopt-go-selfupdate-lib-v1.26.3.md](0008-PLAN-adopt-go-selfupdate-lib-v1.26.3.md).
