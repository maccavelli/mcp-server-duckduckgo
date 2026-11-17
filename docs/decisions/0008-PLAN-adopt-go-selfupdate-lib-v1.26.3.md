---
status: proposed
date: 2026-11-17
associated-madr: "0008-MADR-adopt-go-selfupdate-lib-v1.26.3.md"
---
# Implement go-selfupdate-lib v1.26.3 inside the existing update command

Associated MADR: [0008-MADR-adopt-go-selfupdate-lib-v1.26.3.md](0008-MADR-adopt-go-selfupdate-lib-v1.26.3.md)

## Goal

* `github.com/maccavelli/go-selfupdate-lib v1.26.3` owns the update flag set, the contradiction checks, the run, the build stamps, the GitHub HTTP client, and the publish workflow.
* The Cobra `update` command, `selfupdate.skip-config`, the saved protocol stdout, `main`'s exit mapping, the three-platform standalone installer, and `mcplib v1.6.0` stay.
* `go 1.26.11` and `toolchain go1.27.4` stay.
* The plan stops before any tag, push, or release.

## Scope

In scope, when the owner accepts the MADR and tells the executor to proceed:

* `go.mod` and `go.sum`, limited to adding `v1.26.3` and the tidy that follows.
* `cmd/mcp-server-duckduckgo/update.go`, `update_test.go`, `main.go`, `root.go`, and deletion of `version.go`.
* The four `-X` pairs in the `Makefile`.
* The `uses:` line and `with:` block of the `release` job in `.github/workflows/ci.yml`, and the comment that names the called workflow.
* A dated pointer at the end of [0003-MADR-update-dependency-pins.md](0003-MADR-update-dependency-pins.md).
* `docs/README.md` already indexes this pair. On acceptance, leave the index rows and update their meaning only if a title changes.

Out of scope:

* `selfupdate/service`, a non-zero `HandOff`, `cli.Command`, and the six platforms in the library guide's sample.
* `prerelease-channels-json`, `sign`, and `bridge-release`.
* The `go` line, the `toolchain` line, `mcplib v1.6.0`, fetchup `v0.3.0`, Playwright, stealth, book search, `.golangci.yml`, and the action SHAs for checkout, setup-go, and upload-artifact.
* A tag, a push, or a GitHub release. The next tag the owner asks for is the publisher check named in the MADR.
* Rewriting records 0001, 0002, 0004, 0005, 0006, or 0007. The 0003 edit is the pointer below and nothing else.

## Grounding Reference

The decision, the chosen shape, and evidence E1 to E15 are in [0008-MADR-adopt-go-selfupdate-lib-v1.26.3.md](0008-MADR-adopt-go-selfupdate-lib-v1.26.3.md). Line facts below were read at `035e171de3cd215bde2b4c782e3a31d9de643a0c` and are the lines the edits replace.

* `go.mod:3` is `go 1.26.11`. `go.mod:5` is `toolchain go1.27.4`. `go.mod:11` is `github.com/maccavelli/mcplib v1.6.0`. `go.mod:19` is `golang.org/x/sys v0.50.0`. `go.mod:55` is `golang.org/x/mod v0.43.0`. `go.mod:57` is `golang.org/x/term v0.48.0`. `go.mod:48` is `github.com/ysmood/fetchup v0.3.0`.
* `cmd/mcp-server-duckduckgo/update.go` imports `github.com/maccavelli/mcplib/selfupdate`. `defaultUpdater` takes an `io.Writer`. `GitHubOptions.Client` is `&http.Client{Timeout: updateOperationTimeout}` with `updateOperationTimeout` at 15 minutes. `UserAgent` is `updateRepository.Name + "/" + RawVersion`. `RunE` builds the updater, then a 15-minute child of `cmd.Context()`, then calls `Updater.Run`. Flags are registered with `BoolVar`, `BoolVarP`, and `StringVar` for `check`, `yes`/`y`, `force`, and `version`.
* `cmd/mcp-server-duckduckgo/version.go` holds `RawVersion`, `RawBuildKind`, and `Version`.
* `cmd/mcp-server-duckduckgo/root.go` `Execute` assigns `OriginalStdout = os.Stdout`, then `os.Stdout = os.Stderr`, then `RootCmd.Version = Version`.
* `cmd/mcp-server-duckduckgo/main.go` imports `mcplib/selfupdate` and calls `selfupdate.ExitCode(selfupdate.Result{}, err)`. It prints `err` when the code is not 10.
* `cmd/mcp-server-duckduckgo/serve.go:65` logs `cmd.Version`. `serve.go:121` passes that string to `mcplib.NewMCPServer`. Neither line needs an edit if `RootCmd.Version` stays a single version token with one leading `v` removed.
* `Makefile:34`, `Makefile:40`, `Makefile:44`, and `Makefile:48` each contain `-X main.RawVersion=$(VERSION) -X main.RawBuildKind=$(BUILD_KIND)`.
* `.github/workflows/ci.yml:41` passes `BUILD_KIND=release` for a tag. `.github/workflows/ci.yml:103` is the mcplib workflow pin `8d4d89ff34ff79c1ed65f223cc3d83b3ec0550aa`. `.github/workflows/ci.yml:109` is `bridge-release: false`.
* Parent-module `mcplib` imports that must still compile after the swap are `serve.go`, `serve_test.go`, `internal/util/mcp_helpers.go`, `internal/util/mcp_helpers_test.go`, `internal/handler/system/tools.go`, `internal/handler/system/tools_test.go`, `internal/handler/media/tools.go`, and `internal/handler/search/tools.go`.

Library facts the steps rely on, already measured in the MADR: `Flags.Bind` on a Cobra flag set registers `-y`; `Flags.Request` refuses the three `--check` contradictions before an updater exists; `cli.Run` does not call `os.Exit`; `DefaultHTTPClient().Timeout` is 0; `Summary` is empty for a generic error and non-empty for `ErrUpdateAvailable`; a kind stamp of `release` is `KindRelease` only together with a release tag; `SilenceUsage` and `SilenceErrors` still allow `--help`.

## Prerequisites

* The MADR status is moved to `accepted` in the same implementation, and the owner has told the executor to proceed. Until that, this file stays `proposed` and no product file changes.
* Start from a clean `main`. Re-read the line facts above if `HEAD` is no longer `035e171de3cd215bde2b4c782e3a31d9de643a0c`.
* Edits land through a script or a single-site edit that asserts the old string's count before replacing, then the result is read back.
* A full-module `go test ./...` is a heavy run. On the session host it goes through `~/.agents/skills/available-dev-hosts/scripts/run-heavy -- <command>`. One heavy run at a time. `go test` of `./cmd/mcp-server-duckduckgo` alone may stay on the session host. Darwin-only tests stay on the macOS host.
* The library resolved by `go get` is the published tag `v1.26.3`, whose peeled commit is `1dffdc9dee24a7534513dc51c9039408be37c09b`. A local checkout of the library is not a replace directive.

## Phase 1 - Module

* From the module root, with a clean tree, run the two commands below.

```bash
go get github.com/maccavelli/go-selfupdate-lib@v1.26.3
go mod tidy
```

* Stop if any of these fail. Each one is a deviation, recorded before any workaround.
  * `go.mod` still has exactly one `go 1.26.11` line and one `toolchain go1.27.4` line.
  * `go list -m github.com/maccavelli/go-selfupdate-lib` prints `github.com/maccavelli/go-selfupdate-lib v1.26.3`.
  * `go list -m github.com/maccavelli/mcplib` prints `github.com/maccavelli/mcplib v1.6.0`.
  * `go list -m golang.org/x/sys` prints `golang.org/x/sys v0.50.0`.
  * `go list -m golang.org/x/mod` prints `golang.org/x/mod v0.43.0`.
  * `go list -m golang.org/x/term` prints `golang.org/x/term v0.48.0`.
  * `go list -m github.com/ysmood/fetchup` prints `github.com/ysmood/fetchup v0.3.0`.
  * `go list -m github.com/mxschmitt/playwright-go` prints `github.com/mxschmitt/playwright-go v0.6401.0`.
* `go mod tidy -diff` exits 0 after that tidy. Do not run `go get -u`.

## Phase 2 - Command and stamps

`root.go`:

* Add `protocolStdout io.Writer` next to `OriginalStdout`.
* In `Execute`, after `OriginalStdout = os.Stdout` and before `os.Stdout = os.Stderr`, set `protocolStdout = OriginalStdout`.
* Replace `RootCmd.Version = Version` with `RootCmd.Version = strings.TrimPrefix(buildinfo.Identity().Current(), "v")`.

`main.go`: change the import to `github.com/maccavelli/go-selfupdate-lib/selfupdate`. Leave the `ExitCode` call, the code-10 suppression, and `os.Exit` where they are.

Delete `version.go`.

`update.go` becomes the shape below. Names may match the file's style. Behavior matches this shape.

```go
var newUpdateUpdater = defaultNewUpdateUpdater

var updateOptions = commandUpdateOptions

func defaultNewUpdateUpdater() (*selfupdate.Updater, error) {
    src, err := selfupdate.NewGitHubSource(selfupdate.GitHubOptions{
        Repository: updateRepository,
        Client:     selfupdate.DefaultHTTPClient(),
        UserAgent:  selfupdate.UserAgent(updateRepository.Name, buildinfo.Identity().Current()),
        Limits:     selfupdate.DefaultLimits(),
    })
    if err != nil {
        return nil, err
    }
    selector, err := selfupdate.NewExactAssetSelector(updatePlatforms)
    if err != nil {
        return nil, err
    }
    installer, err := selfupdate.NewStandaloneInstaller(selfupdate.InstallOptions{})
    if err != nil {
        return nil, err
    }
    return selfupdate.New(selfupdate.Config{
        Source:    src,
        Versions:  selfupdate.NewStrictVersionPolicy(),
        Assets:    selector,
        Installer: installer,
        Reporter:  selfupdate.DiscardReporter(),
        Confirmer: selfupdate.NonInteractiveConfirmer(),
        Limits:    selfupdate.DefaultLimits(),
    })
}

func commandUpdateOptions(cmd *cobra.Command, asJSON bool) (cli.Options, error) {
    in := cmd.InOrStdin()
    o := cli.Options{Stderr: cmd.ErrOrStderr(), Stdin: in, JSON: asJSON}
    if in == os.Stdin {
        o.Interactive = cli.StdioOptions().Interactive
    }
    if asJSON {
        if protocolStdout == nil {
            return cli.Options{}, errors.New("update --json: protocol stdout is not preserved")
        }
        o.Stdout = protocolStdout
    }
    return o, nil
}
```

`newUpdateCmd` keeps `Use: "update"`, `Args: cobra.NoArgs`, and the annotation map. Set `SilenceUsage: true` and `SilenceErrors: true` on the command literal. `RunE` does these in order:

* `req, err := flags.Request(updateRepository.Name, buildinfo.Identity())`. Return `err` before constructing an updater.
* `o, err := updateOptions(cmd, flags.JSON)`. Return `err` before constructing an updater.
* `u, err := newUpdateUpdater()`. Return `err`.
* Use `cmd.Context()` when it is non-nil, and `context.Background()` when it is nil.
* `_, err = cli.Run(ctx, u, req, o)`. Return `err`.

After the command literal exists, `flags.Bind(cmd.Flags())`. The `cli.Flags` value must be the one `RunE` closes over, and `Bind` must see the address of that value. Do not also register the same flags by hand.

Delete `updateOperationTimeout`, `buildKind`, `releaseBuildKind`, the `http.Client` timeout, `NewTextReporter`, and `NewTerminalConfirmer`. Keep `updateRepository` and `updatePlatforms` as they are: owner `maccavelli`, name `mcp-server-duckduckgo`, platforms linux/amd64, darwin/arm64, windows/amd64.

`Makefile` recipes `build`, `linux`, `darwin-arm64`, and `windows-amd64`: replace each `-X main.RawVersion=$(VERSION) -X main.RawBuildKind=$(BUILD_KIND)` with:

```text
-X github.com/maccavelli/go-selfupdate-lib/buildinfo.version=$(VERSION) -X github.com/maccavelli/go-selfupdate-lib/buildinfo.kind=$(BUILD_KIND)
```

Leave each recipe's `CGO_ENABLED`, `-trimpath`, `-tags`, `-extldflags`, `-s -w`, `GOOS`, `GOARCH`, and output path as they are. Leave the `BUILD_KIND?=local` default. The comment that cites MADR 0005 stays; that citation names mcplib's build-kind record. Add one sentence to that comment that the linker symbols are `buildinfo.version` and `buildinfo.kind`, and that kind `release` counts only when `VERSION` is a release tag.

## Phase 3 - Publish workflow

In `.github/workflows/ci.yml`, replace the `uses:` value of the `release` job with:

```yaml
uses: maccavelli/go-selfupdate-lib/.github/workflows/publish-selfupdate-release.yml@1dffdc9dee24a7534513dc51c9039408be37c09b # v1.26.3
```

Delete the `bridge-release: false` line. Leave `artifact-name`, `products-json`, `platforms-json`, and `extra-assets-json` as they are, including the three-platform JSON. Do not add `prerelease-channels-json` or `sign`. Leave the job `if:`, the `permissions:` block, and the action pins. Update the comment above the job so it names go-selfupdate-lib at `v1.26.3` and still says why the reusable workflow replaced a floating release action. Leave the `BUILD_KIND=release` branch of the build step as it is.

The annotated tag object `10316bca82c00aeefde4bc614d183d343ed5818e` is the wrong pin. The peeled commit above is the pin. The guide's sample SHA `e62d77cddaf264cc6ec7bbc4f39b90322542f752` is peeled `v1.26.2` and is the wrong pin.

## Phase 4 - Tests

Retarget `update_test.go` at `github.com/maccavelli/go-selfupdate-lib/selfupdate`. The constructor seam is `func() (*selfupdate.Updater, error)`. Tests that only need to prove a flag error replaces it with a function that records a call and returns an error.

Keep, adjusted only for the new seam and import:

* `TestUpdateDoesNotLoadConfig`
* `TestUpdateLeavesNoCacheDirectory`
* `TestOrdinaryCommandStillLoadsConfig`
* `TestUpdateCommandIsAnnotatedToSkipConfig`
* `TestUpdateRejectsPositionalArgs`
* `TestUpdateUsesCallerContext`, still using the real constructor and a cancelled context
* `TestExitCodeMapping`
* `TestUpdatePlatformsAreTheFrozenMatrix`

Change:

* `TestUpdateFlagSurface` also requires `dry-run`, `json`, `channel`, and shorthand `y` on `yes`. It asserts `SilenceUsage` and `SilenceErrors` are true.
* `TestUpdateRejectsContradictoryFlags` runs `--check` with `--yes`, `--force`, and `--dry-run`. The error is non-nil, `errors.Is` against `ErrUpdateAvailable` is false, and the constructor records zero calls.
* `TestUpdateWritesNothingToProtocolStdout` inspects `commandUpdateOptions`. For `--check`, `Stdout` is nil and `Stderr` is the command error stream. The command's own output buffer stays empty.

Add:

* `TestJSONRequiresPreservedStdout`: `--json` while `protocolStdout` is nil returns `update --json: protocol stdout is not preserved` and does not call the constructor. Restore `protocolStdout` with `t.Cleanup`.
* `TestCheckAndJSONStreams`: build an updater with `selfupdatetest.NewFakeSource` and `selfupdatetest.NewRelease` for product `mcp-server-duckduckgo`, tag `v1.2.3`, immutable, covering `updatePlatforms`, and `NewStandaloneInstaller` of an empty `InstallOptions`. Skip when `runtime.GOOS/runtime.GOARCH` is outside `updatePlatforms`. Run `--check` and expect `errors.Is(err, selfupdate.ErrUpdateAvailable)`, a command output buffer of length 0, and an error-stream buffer that contains `mcp-server-duckduckgo: update available:`. Run `--check --json` with `protocolStdout` set to a buffer. That buffer contains `"kind":"result"` and `"exit_code":10`. The command output buffer stays empty. Do not set `Yes`. Do not run an apply. The test binary must still be the process executable after the test.
* `TestUnstampedIdentityIsLocal`: with no extra linker flags, `buildinfo.Identity().Kind` is `buildinfo.KindLocal`, and `NewStrictVersionPolicy().Validate("dev")` returns an error.
* `TestStampCase`: when `SELFUPDATE_STAMP_CASE` is empty, skip. Otherwise require the identity below. A skip while the variable is set is a failure.

| `SELFUPDATE_STAMP_CASE` | Linker version | Linker kind | Required identity |
| :--- | :--- | :--- | :--- |
| `release` | `v9.9.9` | `release` | `KindRelease`, version `v9.9.9` |
| `dirty` | `v9.9.9-dirty` | `release` | `KindLocal`, non-empty `Reason` |
| `capital` | `v9.9.9` | `Release` | `KindLocal` |
| `local-tag` | `v9.9.9` | `local` | `KindLocal`, version `v9.9.9` |

* `TestMakefileStamps`: from the module root, `make -n build linux darwin-arm64 windows-amd64` with `VERSION=v9.9.9` and `BUILD_KIND=release`. The output contains `buildinfo.version=` four times and `buildinfo.kind=` four times, and contains neither `main.RawVersion` nor `main.RawBuildKind`.
* `TestReleaseWorkflowPin`: read `.github/workflows/ci.yml` from the module root. The file contains `maccavelli/go-selfupdate-lib/.github/workflows/publish-selfupdate-release.yml@1dffdc9dee24a7534513dc51c9039408be37c09b`. It does not contain `bridge-release`, `8d4d89ff34ff79c1ed65f223cc3d83b3ec0550aa`, or `prerelease-channels-json`. It still contains the three platform objects and `BUILD_KIND=release`.

Module root in tests is the parent of `cmd/`, found by walking up from the test's working directory until `go.mod` names this module. Do not embed a machine path.

Each new test is first run against a deliberately broken copy, outside the tree: a stamp case whose expected kind is flipped, a workflow copy that still has `bridge-release`, and a makefile copy that still has `main.RawVersion`. Each of those must fail. Record the failure text in the execution record. The tree itself is not the broken copy.

## Phase 5 - Verification

Run these after the last product edit. Capture the exit status before any filter. A non-zero status stops the phase.

Session host, this package:

```bash
gofmt -l cmd/mcp-server-duckduckgo/update.go cmd/mcp-server-duckduckgo/update_test.go cmd/mcp-server-duckduckgo/main.go cmd/mcp-server-duckduckgo/root.go
go test -count=1 ./cmd/mcp-server-duckduckgo
SELFUPDATE_STAMP_CASE=release go test -count=1 -ldflags "-X github.com/maccavelli/go-selfupdate-lib/buildinfo.version=v9.9.9 -X github.com/maccavelli/go-selfupdate-lib/buildinfo.kind=release" ./cmd/mcp-server-duckduckgo -run TestStampCase
SELFUPDATE_STAMP_CASE=dirty go test -count=1 -ldflags "-X github.com/maccavelli/go-selfupdate-lib/buildinfo.version=v9.9.9-dirty -X github.com/maccavelli/go-selfupdate-lib/buildinfo.kind=release" ./cmd/mcp-server-duckduckgo -run TestStampCase
SELFUPDATE_STAMP_CASE=capital go test -count=1 -ldflags "-X github.com/maccavelli/go-selfupdate-lib/buildinfo.version=v9.9.9 -X github.com/maccavelli/go-selfupdate-lib/buildinfo.kind=Release" ./cmd/mcp-server-duckduckgo -run TestStampCase
SELFUPDATE_STAMP_CASE=local-tag go test -count=1 -ldflags "-X github.com/maccavelli/go-selfupdate-lib/buildinfo.version=v9.9.9 -X github.com/maccavelli/go-selfupdate-lib/buildinfo.kind=local" ./cmd/mcp-server-duckduckgo -run TestStampCase
```

`gofmt -l` prints nothing. Each stamp command's output contains `PASS` and does not contain `SKIP`.

Then:

```bash
go vet ./...
make lint
```

`make lint` uses the golangci-lint the CI workflow installs, `v2.15.0`. Do not edit `.golangci.yml` to clear a finding. A new finding stops the phase and goes to the owner.

Heavy run, one at a time:

```bash
~/.agents/skills/available-dev-hosts/scripts/run-heavy -- go test -count=1 -timeout 20m ./...
```

Records, after the 0003 pointer and any status edit:

```bash
make check-records
make markdownlint
```

Search the changed files for a home directory, an account name, and a machine path. A hit is a stop. Describe the search in the execution record without quoting a hit.

`go version -m` on one locally built binary is optional and stays out of the commit. The stamp tests are the identity check.

## Phase 6 - Records and the commit

When implementation starts:

* Set this plan's frontmatter `status` to `in-progress`.
* Set the MADR frontmatter `status` to `accepted`.
* Append this heading and paragraph at the end of `0003-MADR-update-dependency-pins.md`, leaving the deferred paragraph in place:

```markdown
### Amendment 2026-11-17 — self-update library

The separate decision in the last paragraph is
[0008-MADR-adopt-go-selfupdate-lib-v1.26.3.md](0008-MADR-adopt-go-selfupdate-lib-v1.26.3.md).
That paragraph stays.
```

Before `git commit`, `gofmt` every staged `.go` file. Run per-file `golint` on those files when the `golint` binary is on `PATH`. A `golint` finding stops the commit. If `golint` is absent, write that in the execution record and rely on `make lint`. Confirm `git rev-parse --path-format=absolute --git-path hooks` is the fleet global hooks directory or a wrapper that chains to it. Then `git commit --no-edit`. Do not pass a message flag, do not set `GIT_EDITOR=true`, and do not pass `--no-verify`. Do not push. Do not tag.

This plan's status becomes `implemented` only when every completion criterion below has been met and the commit exists. A commit by itself does not make it `implemented`.

## Rollback

Before any tag, revert with new commits. Restore the `mcplib/selfupdate` imports in `update.go`, `update_test.go`, and `main.go`. Restore `version.go`, the `RawVersion` display assignment, and the four `-X main.RawVersion` / `-X main.RawBuildKind` pairs. Restore the workflow pin `8d4d89ff34ff79c1ed65f223cc3d83b3ec0550aa` and `bridge-release: false`. Drop the `go-selfupdate-lib` requirement with `go get` and `go mod tidy`, and confirm `mcplib` is still `v1.6.0` and the `go` and `toolchain` lines are unchanged. Drop the 0003 pointer. Do not rewrite history. A binary already installed is unchanged by the revert; only a later tag would publish either workflow.

## Completion Criteria

* `go list -m github.com/maccavelli/go-selfupdate-lib` is `v1.26.3`, and the Phase 1 version assertions all hold.
* `grep -R 'maccavelli/mcplib/selfupdate' --include='*.go' .` finds nothing, and `grep -R 'github.com/maccavelli/mcplib"' --include='*.go' .` still finds the harness files named in the grounding reference.
* `version.go` is gone. `main.RawVersion` and `main.RawBuildKind` occur nowhere in the `Makefile`.
* `update.go` contains `cli.Flags`, `Flags.Request`, `cli.Run`, `DefaultHTTPClient`, `DiscardReporter`, and `NonInteractiveConfirmer`, and it does not contain `os.Exit`, `cli.Exit`, `NewTextReporter`, or `NewTerminalConfirmer`.
* `.github/workflows/ci.yml` contains the peeled SHA `1dffdc9dee24a7534513dc51c9039408be37c09b` and does not contain `bridge-release`.
* The package tests, the four stamp invocations, `go vet ./...`, `make lint`, and the heavy `go test ./...` exit 0.
* The three planted broken copies failed before the real tests were counted, and the execution record quotes those failures.
* `make check-records` and `make markdownlint` exit 0.
* No tag was created and no push was made by this plan.
* Frontmatter here is `implemented` only after the commit that contains the change. Until the owner says to proceed, it stays `proposed`.

## Execution record

No phase has run. Status is `proposed`. The scratch module cited by the MADR is evidence for the decision, not execution of this plan.
