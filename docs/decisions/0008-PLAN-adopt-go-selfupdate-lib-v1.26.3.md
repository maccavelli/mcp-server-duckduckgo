---
status: implemented
date: 2026-11-17
associated-madr: "0008-MADR-adopt-go-selfupdate-lib-v1.26.3.md"
---
# Implement go-selfupdate-lib v1.26.3 for the update command and the release workflow

Associated MADR: [0008-MADR-adopt-go-selfupdate-lib-v1.26.3.md](0008-MADR-adopt-go-selfupdate-lib-v1.26.3.md)

## Goal

* `github.com/maccavelli/go-selfupdate-lib v1.26.3` owns the update flag set, the contradiction checks, the run, the build stamps, the GitHub HTTP client, the release build, the generated `install.sh` and `install.ps1`, and the publish workflow.
* The Cobra `update` command, a new `identity` command, `selfupdate.skip-config`, the saved protocol stdout, `main`'s exit mapping, the three-platform standalone installer, and `mcplib v1.6.0` stay.
* `go 1.26.11` and `toolchain go1.27.4` stay.
* The plan stops before any tag, push, or release.
* `make verify` is the validate job's quality target. `disclosure-check` calls the `go-shared` script with `--repo`.

## Scope

In scope once the owner tells the executor to proceed. The MADR is already accepted:

* `go.mod` and `go.sum`, limited to adding `v1.26.3` and the tidy that follows.
* `cmd/mcp-server-duckduckgo/update.go`, `update_test.go`, `main.go`, `root.go`, a new `identity.go`, and deletion of `version.go`.
* The four `-X` pairs in the `Makefile`. Local `build` and `linux` keep `-tags netgo` and `-extldflags '-static'`.
* `selfupdate-release.json` at the repository root.
* The `build` and `release` jobs in `.github/workflows/ci.yml`: the hand-rolled build steps are replaced by the two library workflow calls. The `validate` job keeps its checkout, setup-go, module download, and golangci-lint `v2.15.0` install, checks out `go-shared` outside this tree, and runs `make verify`.
* Makefile targets `fmt-check`, `mod-check`, `disclosure-check`, `vuln`, and `verify`, plus `GO_SHARED ?= ../go-shared` and the matching `.PHONY` names.
* A dated pointer at the end of [0003-MADR-update-dependency-pins.md](0003-MADR-update-dependency-pins.md).
* `docs/README.md` already indexes this pair. Acceptance on 2026-11-17 did not change a title, so the index rows stay.

Out of scope:

* `selfupdate/service`, a non-zero `HandOff`, `cli.Command`, the six platforms in the library guide's sample, and `windows/arm64` on its own (MADR E18).
* Installer hooks, a prerelease channel, `sign`, and `bridge-release`.
* A hand-written `install.sh` or `install.ps1`, and install instructions in the product README.
* A docs workflow, path filters, a concurrency group, `FORCE_JAVASCRIPT_ACTIONS_TO_NODE24`, and a native operating-system test matrix.
* A copy of the disclosure script in this repository. Accepting or editing `go-shared` 0003. Coverage, `scripts/verify-scripts.sh`, actionlint, `check_ci_filters.py`, `tools-test`, and `build-all` as `verify` dependencies. A `tool` directive or a `.tools` bootstrap for govulncheck.
* An `actions/checkout` step whose `path` is `${{ runner.temp }}/go-shared`. The pinned action rejects that path (MADR E21).
* `tags` on the release spec. Published binaries follow the library recipe in MADR E16.
* The `go` line, the `toolchain` line, `mcplib v1.6.0`, fetchup `v0.3.0`, Playwright, stealth, book search, `.golangci.yml`, and the action SHAs for checkout, setup-go, and upload-artifact.
* A tag, a push, or a GitHub release. The next tag the owner asks for is the publisher check named in the MADR.
* Rewriting records 0001, 0002, 0004, 0005, 0006, or 0007. The 0003 edit is the pointer below and nothing else.

## Grounding Reference

The decision, the chosen shape, and evidence E1 to E21 are in [0008-MADR-adopt-go-selfupdate-lib-v1.26.3.md](0008-MADR-adopt-go-selfupdate-lib-v1.26.3.md). Line facts below were read at `035e171de3cd215bde2b4c782e3a31d9de643a0c` and are the lines the edits replace. Commit `1335d2f7f144e30567f233a4b667bba5216e99d5` changes only the docs that introduced the first draft of this pair. E16 was measured the same day against peeled `v1.26.3` and is the installer contract the spec and the identity command satisfy. E17 and E18 are why the workflow copy stops at the two release workflows and why the platform list stays at three. E19 is why `disclosure-check` calls `go-shared` and why `verify` is the subset named in the MADR. E20 is why `vuln` runs `govulncheck` `v1.10.0`. E21 is why the validate job fetches `go-shared` with git into `runner.temp`.

* `go.mod:3` is `go 1.26.11`. `go.mod:5` is `toolchain go1.27.4`. `go.mod:11` is `github.com/maccavelli/mcplib v1.6.0`. `go.mod:19` is `golang.org/x/sys v0.50.0`. `go.mod:55` is `golang.org/x/mod v0.43.0`. `go.mod:57` is `golang.org/x/term v0.48.0`. `go.mod:48` is `github.com/ysmood/fetchup v0.3.0`.
* `cmd/mcp-server-duckduckgo/update.go` imports `github.com/maccavelli/mcplib/selfupdate`. `defaultUpdater` takes an `io.Writer`. `GitHubOptions.Client` is `&http.Client{Timeout: updateOperationTimeout}` with `updateOperationTimeout` at 15 minutes. `UserAgent` is `updateRepository.Name + "/" + RawVersion`. `RunE` builds the updater, then a 15-minute child of `cmd.Context()`, then calls `Updater.Run`. Flags are registered with `BoolVar`, `BoolVarP`, and `StringVar` for `check`, `yes`/`y`, `force`, and `version`.
* `cmd/mcp-server-duckduckgo/version.go` holds `RawVersion`, `RawBuildKind`, and `Version`.
* `cmd/mcp-server-duckduckgo/root.go` `Execute` assigns `OriginalStdout = os.Stdout`, then `os.Stdout = os.Stderr`, then `RootCmd.Version = Version`.
* `cmd/mcp-server-duckduckgo/main.go` imports `mcplib/selfupdate` and calls `selfupdate.ExitCode(selfupdate.Result{}, err)`. It prints `err` when the code is not 10.
* `cmd/mcp-server-duckduckgo/serve.go:65` logs `cmd.Version`. `serve.go:73` passes that string into `run`. `serve.go:121` passes it to `mcplib.NewMCPServer`. None of those lines needs an edit if `RootCmd.Version` stays a single version token with one leading `v` removed.
* `Makefile:34`, `Makefile:40`, `Makefile:44`, and `Makefile:48` each contain `-X main.RawVersion=$(VERSION) -X main.RawBuildKind=$(BUILD_KIND)`.
* `.github/workflows/ci.yml` job `build` starts at the `needs: validate` job and hand-rolls `make build-all`, SHA256SUMS, a linux `--version` check, and `actions/upload-artifact`. `.github/workflows/ci.yml:41` passes `BUILD_KIND=release` for a tag. `.github/workflows/ci.yml:103` is the mcplib workflow pin `8d4d89ff34ff79c1ed65f223cc3d83b3ec0550aa`. `.github/workflows/ci.yml:109` is `bridge-release: false`. The `validate` job runs `go test ./...`, then `go vet ./...`, then one step that installs golangci-lint `v2.15.0` and runs `make lint`. `Makefile:56` is `go fmt ./...`. `Makefile:29` is the `.PHONY` line. There is no `GO_SHARED` variable.
* `Makefile:34` and `Makefile:40` pass `-tags netgo` and `-extldflags '-static'` with `CGO_ENABLED=0`. `Makefile:44` and `Makefile:48` do not set `CGO_ENABLED`.
* `root.go` `init` calls `RootCmd.AddCommand(newUpdateCmd())`. There is no `identity` command.
* Parent-module `mcplib` imports that must still compile after the swap are `serve.go`, `serve_test.go`, `internal/util/mcp_helpers.go`, `internal/util/mcp_helpers_test.go`, `internal/handler/system/tools.go`, `internal/handler/system/tools_test.go`, `internal/handler/media/tools.go`, and `internal/handler/search/tools.go`.
* `update_test.go` `newTestRoot` does not call `Execute`. It replaces `newUpdater` with `func(io.Writer)`, adds only `newUpdateCmd`, and leaves `OriginalStdout` nil. `TestBuildKindMapping` calls `buildKind` and assigns `RawBuildKind`. `TestVersionStampIsEffective` reads `RawVersion`, `RawBuildKind`, and `Version`. `root_test.go` asserts `Use` and `Short` only.
* `selfupdate.ExitCode` uses `errors.Is` for `ErrUpdateAvailable` (`selfupdate/errors.go`). `cli.Flags.Bind` is a pointer receiver and `Flags.Request` is a value receiver (`selfupdate/cli/flags.go`). `buildinfo.Identity` decides kind from the two stamps alone. `Info.String` appends a 12-character revision, and `-dirty` when `vcs.modified` is true (`buildinfo/buildinfo.go`). `releasespec.Spec.InstallerEnvPrefix` and `ExtraNames` return an error as their second result.
* The session host's `make --version` on 2026-11-17 printed `GNU Make 3.81`. A recipe `false | true` under this Makefile's `.SHELLFLAGS` exited 0, so 3.81 stores `.SHELLFLAGS` and does not pass it to the shell. New recipes fail through their own command status.
* `actions/checkout` at `3d3c42e5aac5ba805825da76410c181273ba90b1` throws when `path` resolves outside `GITHUB_WORKSPACE` (MADR E21).

Library facts the steps rely on, already measured in the MADR: `Flags.Bind` on a Cobra flag set registers `-y`; `Flags.Request` refuses the three `--check` contradictions before an updater exists; `cli.Run` does not call `os.Exit`; `DefaultHTTPClient().Timeout` is 0; `Summary` is empty for a generic error and non-empty for `ErrUpdateAvailable`; a kind stamp of `release` is `KindRelease` only together with a release tag; `SilenceUsage` and `SilenceErrors` still allow `--help`.

## Prerequisites

* The MADR frontmatter is already `accepted`. This file stays `proposed` until the owner tells the executor to proceed. Until that word, no product file changes: no Go source, no Makefile, no workflow, no spec, and no edit to `0003-MADR-update-dependency-pins.md`.
* Accepting this MADR leaves `go-shared` `docs/decisions/0003-MADR-house-disclosure-check.md` proposed. Phase 4 stops until that plan is implemented and `scripts/check_disclosure.py` supports `--repo` and `--self-test`. The script stays in `go-shared`. This plan does not vendor it and does not invent the fetch hash.
* The accepted MADR and this plan may already be uncommitted. That is the expected tree. Stop when a product file differs from the line facts above. Re-read those lines when `HEAD` is no longer `035e171de3cd215bde2b4c782e3a31d9de643a0c` for the product sources. `1335d2f7f144e30567f233a4b667bba5216e99d5` is docs on top of that product tree.
* Edits land through a script or a single-site edit that asserts the old string's count before replacing, then the result is read back.
* A full-module `go test ./...` is a heavy run. On the session host it goes through `~/.agents/skills/available-dev-hosts/scripts/run-heavy -- <command>`. `run-heavy` queues on the session host. It does not copy the tree to another machine. One heavy run at a time. `go test` of `./cmd/mcp-server-duckduckgo` alone may stay on the session host. Darwin-only tests stay on the macOS host. `run-on-host` is the wrong wrapper for Phase 6: the copy it makes has no sibling `../go-shared`.
* The library resolved by `go get` is the published tag `v1.26.3`, whose peeled commit is `1dffdc9dee24a7534513dc51c9039408be37c09b`. A local checkout of the library is not a replace directive.
* `make` must be GNU make. On 2026-11-17 the session host's `make --version` was GNU Make 3.81, and that binary does not apply `.SHELLFLAGS`. When `make --version` does not say `GNU Make`, stop. BSD make cannot parse this Makefile's `ifeq`.
* No phase is green by itself. `go mod tidy` drops a requirement no file imports, so Phase 1 runs after the Phase 2 import exists. The package does not compile between that import and the Phase 5 test edits. The workflow and Makefile tests in Phase 5 require Phases 3 and 4. The one commit is Phase 7, after Phase 6 has exited 0. An earlier commit would record a red tree.

## Execution order

The phase headings stay where they are. Run the work in this order.

* When the owner says to proceed, set this plan's frontmatter `status` to `in-progress`. The MADR is already `accepted`. Stop when it is anything else.
* Phase 2 production edits: `update.go`, `root.go`, `main.go`, `identity.go`, deletion of `version.go`, and the four Makefile `-X` pairs. Leave `update_test.go` untouched in this step. Run no `go test`.
* Phase 1, the two module commands and the version assertions. The new import is what keeps `go mod tidy` from dropping `v1.26.3`.
* Phase 3, the spec and the `build` and `release` jobs.
* Phase 4. When its stop conditions fail, leave Phases 2, 1, and 3 uncommitted and stop. Do not write a hash. Do not vendor the script.
* Phase 5, which is what makes `./cmd/mcp-server-duckduckgo` compile again.
* Phase 6.
* Phase 7, then the one `git commit --no-edit`. Do not push. Do not tag.

## Phase 1 - Module

Run this phase at the point Execution order names, after Phase 2 has added an import of `github.com/maccavelli/go-selfupdate-lib` under `cmd/mcp-server-duckduckgo`. `go mod tidy` before that import removes the requirement, and that result is a failed phase.

* From the module root, run the two commands below.

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
* `go mod tidy -diff` exits 0 after that tidy. Do not run `go get -u`. `update_test.go` still imports `mcplib/selfupdate` until Phase 5, so this tidy keeps `mcplib v1.6.0` and the new library together. Run `go mod tidy` again at the end of Phase 5 and require `go mod tidy -diff` to exit 0 there too.

## Phase 2 - Command and stamps

Run the production edits here and leave `update_test.go` unchanged until Phase 5. After this phase the package does not compile: `version.go` is gone, `newUpdater` is gone, and the tests still name both. That window is expected. Run no `go test` and make no commit in the window.

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
* Use `cmd.Context()` when it is non-nil, and `context.Background()` when it is nil. `ExecuteContext` sets a context. The nil branch covers a direct `RunE` call.
* `_, err = cli.Run(ctx, u, req, o)`. Return `err`. `main` passes that error to `selfupdate.ExitCode`, which unwraps with `errors.Is`. A wrapped `ErrUpdateAvailable` still maps to 10. Do not unwrap it in `RunE` or in `main`.

After the command literal exists, `flags.Bind(cmd.Flags())`. Declare `var flags cli.Flags` in `newUpdateCmd`, close over that variable from `RunE`, and call `flags.Bind`, which takes a pointer. `Request` is a value receiver, so it copies the struct at the call, which is after Cobra has parsed. A copy taken before `RunE` would still be the zero flags. Do not also register the same flags by hand. Bind's `--version` is the target-release string on the `update` command. Leave `RootCmd.Version` as the trimmed display string from the step above. Do not set a `Version` field on the update command, and do not keep the hand-rolled `StringVar`.

`cli.Run` with a nil `Signals` slice installs interrupt and `SIGTERM` for the run and removes them when it returns. Leave `Signals` nil. Tests do not set an empty slice to quiet that notifier. `Timeout` stays 0 so `cli.Run` applies `DefaultTimeout`. `HandOff` stays the zero value. Do not call `cli.Exit` or `os.Exit`.

Delete `updateOperationTimeout`, `buildKind`, `releaseBuildKind`, the `http.Client` timeout, `NewTextReporter`, and `NewTerminalConfirmer`. Keep `updateRepository` and `updatePlatforms` as they are: owner `maccavelli`, name `mcp-server-duckduckgo`, platforms linux/amd64, darwin/arm64, windows/amd64.

`Makefile` recipes `build`, `linux`, `darwin-arm64`, and `windows-amd64`: replace each `-X main.RawVersion=$(VERSION) -X main.RawBuildKind=$(BUILD_KIND)` with:

```text
-X github.com/maccavelli/go-selfupdate-lib/buildinfo.version=$(VERSION) -X github.com/maccavelli/go-selfupdate-lib/buildinfo.kind=$(BUILD_KIND)
```

Leave each recipe's `CGO_ENABLED`, `-trimpath`, `-tags`, `-extldflags`, `-s -w`, `GOOS`, `GOARCH`, and output path as they are. Leave the `BUILD_KIND?=local` default. The comment that cites MADR 0005 stays; that citation names mcplib's build-kind record. Add one sentence to that comment that the linker symbols are `buildinfo.version` and `buildinfo.kind`, and that kind `release` counts only when `VERSION` is a release tag. CI does not call these recipes. A published binary is the library recipe in MADR E16.

`identity.go` is the command below. `root.go` `init` registers it with `RootCmd.AddCommand(newIdentityCmd())` next to `newUpdateCmd()`.

```go
func newIdentityCmd() *cobra.Command {
    return &cobra.Command{
        Use:           "identity",
        Args:          cobra.NoArgs,
        SilenceUsage:  true,
        SilenceErrors: true,
        Annotations:   map[string]string{skipConfigAnnotation: skipConfigValue},
        RunE: func(*cobra.Command, []string) error {
            if OriginalStdout == nil {
                return errors.New("identity: protocol stdout is not preserved")
            }
            _, err := fmt.Fprintln(OriginalStdout, buildinfo.Identity().String())
            return err
        },
    }
}
```

`RunE` writes the saved file, which is process stdout. It does not call `fmt.Println`. After `Execute`, `os.Stdout` is the aliased stderr, and a print there fails the identity job (MADR E3, E16). `Identity().String()` appends a space and the first 12 characters of the revision when the build recorded one, and then `-dirty` when `vcs.modified` is true. Kind stays `KindRelease` on a dirty tree when the two stamps are a release tag and `release`. The library identity job rejects a `-dirty` suffix. A clean tag checkout is what makes the published line match. The identity test compares the pipe with `Identity().String()`, suffix included.

## Phase 3 - Release spec and workflows

Write `selfupdate-release.json` at the repository root:

```json
{
  "schema": 1,
  "products": [
    {
      "name": "mcp-server-duckduckgo",
      "package": "./cmd/mcp-server-duckduckgo",
      "identity_args": ["identity"]
    }
  ],
  "platforms": [
    {"os": "linux", "arch": "amd64"},
    {"os": "darwin", "arch": "arm64"},
    {"os": "windows", "arch": "amd64"}
  ],
  "packaging": "binary",
  "installer": {"name": "mcp-server-duckduckgo"}
}
```

That file is the whole spec. It has no `hooks`, no `extras`, no `prerelease_channels`, and no `tags`. `installer.name` is the repository name. The derived prefix is `MCP_SERVER_DUCKDUCKGO` (MADR E16).

In `.github/workflows/ci.yml`, leave the workflow `on:` block as it is. Leave the `validate` job's consumer checkout, setup-go, and `go mod download` as they are. Phase 4 replaces the rest of that job. Replace the `build` job and the `release` job with:

```yaml
  # Build and stage from selfupdate-release.json. A tag stamps the tag as
  # kind release. Any other ref rehearses, and the release job is skipped.
  # The called workflow generates install.sh and install.ps1 (0008).
  build:
    name: Build release
    needs: validate
    permissions:
      contents: read
    uses: maccavelli/go-selfupdate-lib/.github/workflows/build-selfupdate-release.yml@1dffdc9dee24a7534513dc51c9039408be37c09b # v1.26.3
    with:
      spec-path: selfupdate-release.json

  # Publish the staged artifact. extra-assets-json carries install.sh and
  # install.ps1. The job that replaced softprops/action-gh-release still
  # validates, attests, and publishes once (MADR 0005, 0008).
  release:
    name: Publish self-update release
    needs: build
    if: needs.build.outputs.rehearsal == 'false'
    permissions:
      contents: write
      id-token: write
      attestations: write
    uses: maccavelli/go-selfupdate-lib/.github/workflows/publish-selfupdate-release.yml@1dffdc9dee24a7534513dc51c9039408be37c09b # v1.26.3
    with:
      artifact-name: ${{ needs.build.outputs.artifact-name }}
      products-json: ${{ needs.build.outputs.products-json }}
      platforms-json: ${{ needs.build.outputs.platforms-json }}
      extra-assets-json: ${{ needs.build.outputs.extra-assets-json }}
      prerelease-channels-json: ${{ needs.build.outputs.prerelease-channels-json }}
```

Leave `module-dir`, `artifact-name`, `extras-artifact-name`, `retention-days`, and `sign` unset. Do not pass `bridge-release`. The `validate` job's action pins and the golangci-lint `v2.15.0` install stay.

The annotated tag object `10316bca82c00aeefde4bc614d183d343ed5818e` is the wrong pin. The peeled commit above is the pin. The guide's sample SHA `e62d77cddaf264cc6ec7bbc4f39b90322542f752` is peeled `v1.26.2` and is the wrong pin. `7c677842dface4b14b65b2e8b46aa0a7543bed13` is peeled `v1.26.0`. Those two workflow files match the `v1.26.3` peel byte for byte, and this plan still pins the `v1.26.3` peel because that is the release `go.mod` will require.

## Phase 4 - Disclosure check and verify

Run this phase after Phase 3 and before Phase 5. Phase 5's Makefile and workflow assertions pass only once this phase has landed. Phases 1, 2, and 3 do not need the script. When the stop conditions below fail, leave those earlier edits uncommitted and do not start Phase 5.

Stop, and leave the tree unchanged for this phase, when any of these fail. Run them from the module root.

* `go-shared` `scripts/check_disclosure.py` is absent from the sibling checkout at `../go-shared`.
* `python3 ../go-shared/scripts/check_disclosure.py --self-test` exits non-zero.
* `sha=$(git -C ../go-shared rev-parse HEAD)` names a commit that does not contain `scripts/check_disclosure.py`. The check is `git -C ../go-shared cat-file -e "${sha}:scripts/check_disclosure.py"`.
* A scratch directory outside this repository cannot fetch that commit from the public remote. The probe is `git init`, `git remote add origin https://github.com/maccavelli/go-shared.git`, and `git fetch --depth 1 origin "$sha"`. Delete the scratch directory afterwards. A local-only commit fails this probe. Do not put a token on the remote. If git refuses the scratch directory as unsafe, add that directory as `safe.directory` for the probe and record the message. A depth-1 fetch that the remote rejects is a stop. Do not switch to a full clone in the same step.

When those four succeed, `sha` is the value of `GO_SHARED_SHA` in the workflow below. Write that full hash in the execution record. Do not reserve a hash in this file before that commit exists. The hash is 40 lowercase hexadecimal characters. It is distinct from the library peel `1dffdc9dee24a7534513dc51c9039408be37c09b`, the annotated tag object `10316bca82c00aeefde4bc614d183d343ed5818e`, and the mcplib pin `8d4d89ff34ff79c1ed65f223cc3d83b3ec0550aa`.

Add `GO_SHARED ?= ../go-shared` next to `GOLANGCI_LINT` in the Makefile. Add `fmt-check`, `mod-check`, `disclosure-check`, `vuln`, and `verify` to the `.PHONY` line. In the `lint` recipe's install hint, replace `@latest` with `@v2.15.0`. The hint is the line a missing binary prints. `0005` already pinned the binary CI installs, and `verify` runs that `lint` recipe.

Each recipe line in the Makefile begins with one tab. The fence spells those lines with spaces so this record contains none. `$$` in the fence is one `$` after make reads the file.

```makefile
fmt-check: ## Checks Go formatting without rewriting files
  @if ! git ls-files -- '*.go' | grep -q .; then \
    echo "fmt-check: no tracked Go files"; \
    exit 1; \
  fi; \
  unformatted="$$(git ls-files -z -- '*.go' | xargs -0 gofmt -l)"; \
  status=$$?; \
  if [ "$$status" -ne 0 ]; then \
    echo "gofmt failed"; \
    exit "$$status"; \
  fi; \
  if [ -n "$$unformatted" ]; then \
    echo "gofmt found unformatted files:"; \
    printf '%s\n' "$$unformatted"; \
    exit 1; \
  fi

mod-check: ## Verifies module tidiness and downloaded module checksums
  go mod tidy -diff
  go mod verify

disclosure-check: ## Scans tracked text for real user-home paths (go-shared 0003)
  python3 "$(GO_SHARED)/scripts/check_disclosure.py" --repo .

vuln: ## Reports reachable vulnerabilities with govulncheck v1.10.0
  go run golang.org/x/vuln/cmd/govulncheck@v1.10.0 ./...

verify: mod-check fmt-check lint vet test vuln check-records markdownlint disclosure-check ## Runs the local and CI quality checks
```

The empty check exits before `xargs`. An empty `gofmt -l` argument list reads stdin, and macOS `xargs` runs the command when its input is empty. GNU `xargs` does the same unless `-r` is passed, and macOS `xargs` has no `-r`, so the recipe does not use `-r`. The names stay on the pipe. A shell variable would keep nothing after the first NUL. The recipe records the command-substitution status because GNU Make 3.81 does not apply `.SHELLFLAGS`, and an assignment does not abort the recipe by itself. A tracked Go path that contains a newline is a stop if one ever appears. No tracked Go path in this repository contains a space or a newline. `git ls-files` lists the index. `identity.go` is invisible to `fmt-check` and to `disclosure-check` until it is staged. Phase 6 stages the commit set before `make verify`. The existing `fmt` target stays, and `verify` does not depend on it, because `go fmt` writes. `fmt-check` does not call `golangci-lint fmt`. This repository has no `tools` target to supply that binary. `vuln` uses `go run` at `v1.10.0`, so `go.mod` gains no `tool` directive and the recipe does not look up a binary on `PATH` (MADR E20). The 2026-11-17 run left `go.mod` and `go.sum` unchanged. Do not add `GOFLAGS=-mod=readonly` to force a tool directive. The scan needs the network for that module and for the vulnerability database. An unreachable database fails the phase. Do not add a skip or `continue-on-error`. `verify` has no recipe of its own. GNU make runs those dependencies from left to right when `-j` is omitted. Phase 6 and the validate job omit `-j`. Coverage, `verify-scripts.sh`, actionlint, `check_ci_filters.py`, `tools-test`, and `build-all` are not dependencies. `markdownlint` runs `npx`, so `verify` also needs the network for that package.

In the `validate` job, keep the consumer checkout, setup-go, and `go mod download`. Delete the `go test ./...` step and the `go vet ./...` step. Replace the install-and-lint step with the fetch, the install, and `make verify`. The install stays in front of `make verify` because `lint` needs that binary. The consumer checkout pin stays `actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1`.

```yaml
      - name: Fetch go-shared
        env:
          GO_SHARED_SHA: <the 40-character sha from the stop check>
        run: |
          set -euo pipefail
          dest="${{ runner.temp }}/go-shared"
          rm -rf "${dest}"
          git init "${dest}"
          git -C "${dest}" remote add origin https://github.com/maccavelli/go-shared.git
          git -C "${dest}" fetch --depth 1 origin "${GO_SHARED_SHA}"
          git -C "${dest}" checkout --detach FETCH_HEAD
          test "$(git -C "${dest}" rev-parse HEAD)" = "${GO_SHARED_SHA}"
      - run: |
          go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0
          GOLANGCI_LINT="$(go env GOPATH)/bin/golangci-lint" make verify GO_SHARED="${{ runner.temp }}/go-shared"
```

The committed workflow replaces the angle-bracket token with the 40-character `sha`. The token in this plan is the instruction. The fetch uses no token. `persist-credentials` does not apply, because this step is not `actions/checkout`. Do not add a second `actions/checkout` with `path: ${{ runner.temp }}/go-shared`. That pin throws when the path is outside `GITHUB_WORKSPACE` (MADR E21). `${{ runner.temp }}` is outside the workspace, so `make markdownlint`'s glob walks this repository only. A scratch probe showed `git ls-files` of markdown omitting files inside a nested git directory while `git status` still listed that directory (MADR E19). The fetch stays outside either way. The default `GO_SHARED` of `../go-shared` is what a local run uses. CI passes the runner path on the `make` command line, which overrides `?=`. `setup-go` and `go mod download` stay between the consumer checkout and this fetch. Leave `GOBIN` unset so `go install` and the `lint` recipe share `$(go env GOPATH)/bin`. The explicit `GOLANGCI_LINT=` assignment is the same one the current install step uses.

## Phase 5 - Tests

Retarget `update_test.go` at `github.com/maccavelli/go-selfupdate-lib/selfupdate`. The constructor seam is `newUpdateUpdater` with type `func() (*selfupdate.Updater, error)`. `newTestRoot` takes that function. `unreachableBuild` drops the `io.Writer` argument. Tests that only need to prove a flag error pass a function that increments a counter and returns an error. They do not pass `defaultNewUpdateUpdater`.

`newTestRoot` still does not call `Execute`, still adds only `newUpdateCmd`, and still leaves `OriginalStdout` and `protocolStdout` nil. A test that calls `ExecuteContext` never runs the stdout alias, so it sets whichever of those variables it needs and restores both with `t.Cleanup`. Do not add `t.Parallel` to a test that writes `newUpdateUpdater`, `loadConfig`, `OriginalStdout`, or `protocolStdout`. `root_test.go` needs no edit.

Delete `TestBuildKindMapping` and `TestVersionStampIsEffective`. Both read `RawBuildKind`, `RawVersion`, `Version`, or `buildKind`, and those names leave with `version.go`.

Keep, adjusted only for the new seam and import:

* `TestUpdateDoesNotLoadConfig`
* `TestUpdateLeavesNoCacheDirectory`
* `TestOrdinaryCommandStillLoadsConfig`
* `TestUpdateCommandIsAnnotatedToSkipConfig`
* `TestUpdateRejectsPositionalArgs`
* `TestUpdateUsesCallerContext`, still using `defaultNewUpdateUpdater` and a cancelled context. The error must satisfy `errors.Is(err, context.Canceled)`. `NewGitHubSource` does not take that context and did not dial in MADR E10, so a constructor error is a different failure. Stop and record it. Do not loosen the assertion. Deviation 2026-11-17: the test re-executes from a copy of the test binary under the home directory before that assertion. The assertion stays.
* `TestExitCodeMapping`
* `TestUpdatePlatformsAreTheFrozenMatrix`

Change:

* `TestUpdateFlagSurface` also requires `dry-run`, `json`, `channel`, and shorthand `y` on `yes`. It asserts `SilenceUsage` and `SilenceErrors` are true. `Bind` has already registered those names on the command `newUpdateCmd` returns.
* `TestUpdateRejectsContradictoryFlags` runs `--check` with `--yes`, `--force`, and `--dry-run` through the recording seam. The error is non-nil, `errors.Is` against `ErrUpdateAvailable` is false, and the counter stays 0. `Flags.Request` returns before `newUpdateUpdater`.
* `TestUpdateWritesNothingToProtocolStdout` calls `commandUpdateOptions` on a command whose `SetOut` and `SetErr` are different buffers. For `--check`, `Stdout` is nil and `Stderr` is the command error stream. The output buffer stays empty. The old seam passed the writer into the constructor. This test has no writer argument to compare.

Add:

* `TestJSONRequiresPreservedStdout`: `--json` while `protocolStdout` is nil returns `update --json: protocol stdout is not preserved` and does not call the constructor. `commandUpdateOptions` returns that error before `newUpdateUpdater`. Restore `protocolStdout` with `t.Cleanup`.
* `TestCheckAndJSONStreams`: replace `newUpdateUpdater` with a function that returns `selfupdate.New` whose `Source` is `selfupdatetest.NewFakeSource`. The release comes from `selfupdatetest.NewRelease` for product `mcp-server-duckduckgo`, tag `v1.2.3`, immutable, platforms `updatePlatforms`, and a short body per platform. The updater uses `NewStrictVersionPolicy`, `NewExactAssetSelector(updatePlatforms)`, `NewStandaloneInstaller` of an empty `InstallOptions`, `DiscardReporter`, and `NonInteractiveConfirmer`. Skip when `runtime.GOOS/runtime.GOARCH` is outside `updatePlatforms`. Use three buffers: command output, command error, and `protocolStdout`. `newTestRoot` points output and error at one buffer, so this test builds the command itself. Run `--check` and expect `errors.Is(err, selfupdate.ErrUpdateAvailable)`, an output buffer of length 0, and an error buffer that contains `mcp-server-duckduckgo: update available:` and `local build: apply requires --force`. The unstamped test binary is `KindLocal`, and `updater.go` adds that detail for `CheckOnly` with `OperationReplaceLocal`. Run `--check --json` with `protocolStdout` set to its buffer. That buffer contains `"kind":"result"` and `"exit_code":10`. The output buffer stays empty. Do not set `Yes`. Do not run an apply. `CheckOnly` returns before apply. The test binary must still be the process executable after the test. Restore `protocolStdout` and the seam with `t.Cleanup`. Deviation 2026-11-17: the test re-executes from a copy of the test binary under the home directory before those assertions. The assertions stay. Empty `InstallOptions` stays.
* `TestUnstampedIdentityIsLocal`: with no extra linker flags, `buildinfo.Identity().Kind` is `buildinfo.KindLocal` and `Current()` is `dev`. `NewStrictVersionPolicy().Validate(Current())` returns an error. A `Current()` other than `dev` is a stop: that is not the unstamped `go test` case in MADR E11. A dirty working tree does not flip `Kind`.
* `TestStampCase`: when `SELFUPDATE_STAMP_CASE` is empty, skip. Otherwise require the identity below. A skip while the variable is set is a failure.

| `SELFUPDATE_STAMP_CASE` | Linker version | Linker kind | Required identity |
| :--- | :--- | :--- | :--- |
| `release` | `v9.9.9` | `release` | `KindRelease`, version `v9.9.9` |
| `dirty` | `v9.9.9-dirty` | `release` | `KindLocal`, non-empty `Reason` |
| `capital` | `v9.9.9` | `Release` | `KindLocal` |
| `local-tag` | `v9.9.9` | `local` | `KindLocal`, version `v9.9.9` |

* `TestMakefileStamps`: from the module root, GNU `make -n build linux darwin-arm64 windows-amd64` with `VERSION=v9.9.9` and `BUILD_KIND=release`. The output contains `buildinfo.version=` four times and `buildinfo.kind=` four times, and contains neither `main.RawVersion` nor `main.RawBuildKind`. The `build` and `linux` lines still contain `-tags netgo` and `-extldflags '-static'`.
* `TestIdentityPrintsBuildinfo`: `newIdentityCmd` has the skip-config annotation, `SilenceUsage`, and `SilenceErrors`. With `loadConfig` counting calls, an `os.Pipe` assigned to `OriginalStdout`, and a root whose `PersistentPreRunE` is `RootCmd.PersistentPreRunE`, `identity` returns nil, `loadConfig` stays at 0, and the pipe reads exactly `buildinfo.Identity().String()` plus one newline. The command output buffer stays empty. A nil `OriginalStdout` returns `identity: protocol stdout is not preserved`. Restore `OriginalStdout` with `t.Cleanup`. `RootCmd` lists a command whose `Use` is `identity`.
* `TestReleaseSpec`: walk to the module root and `releasespec.Parse` `selfupdate-release.json`. The spec has one product, name `mcp-server-duckduckgo`, package `./cmd/mcp-server-duckduckgo`, and `identity_args` equal to `identity`. Platforms are linux/amd64, darwin/arm64, and windows/amd64, in that order. Packaging is `binary`. `InstallerName("mcp-server-duckduckgo")` is `mcp-server-duckduckgo`. `InstallerEnvPrefix("mcp-server-duckduckgo")` returns `MCP_SERVER_DUCKDUCKGO` and a nil error. `InstallerScripts` is `install.sh` then `install.ps1`. `ExtraNames("v9.9.9")` returns that same pair and a nil error. `Installer.Hooks`, `Extras`, and `PrereleaseChannels` are empty. The product `Tags` are empty.
* `TestReleaseWorkflowPin`: read `.github/workflows/ci.yml` from the module root. The file contains both `maccavelli/go-selfupdate-lib/.github/workflows/build-selfupdate-release.yml@1dffdc9dee24a7534513dc51c9039408be37c09b` and `maccavelli/go-selfupdate-lib/.github/workflows/publish-selfupdate-release.yml@1dffdc9dee24a7534513dc51c9039408be37c09b`, `spec-path: selfupdate-release.json`, and `needs.build.outputs.rehearsal == 'false'`. It contains each of `needs.build.outputs.artifact-name`, `needs.build.outputs.products-json`, `needs.build.outputs.platforms-json`, `needs.build.outputs.extra-assets-json`, and `needs.build.outputs.prerelease-channels-json`. It contains `https://github.com/maccavelli/go-shared.git`, `fetch --depth 1 origin`, `checkout --detach FETCH_HEAD`, `golangci-lint@v2.15.0`, `GOLANGCI_LINT="$(go env GOPATH)/bin/golangci-lint"`, `${{ runner.temp }}/go-shared`, and `make verify GO_SHARED=`. The `GO_SHARED_SHA` value is 40 lowercase hexadecimal characters. The test rejects that value when it equals `1dffdc9dee24a7534513dc51c9039408be37c09b`, `10316bca82c00aeefde4bc614d183d343ed5818e`, or `8d4d89ff34ff79c1ed65f223cc3d83b3ec0550aa`. The test does not open this plan. The executor shows, in the execution record, that the same value is `git -C ../go-shared rev-parse HEAD`. The file does not contain `path: ${{ runner.temp }}/go-shared`, `repository: maccavelli/go-shared`, `bridge-release`, `8d4d89ff34ff79c1ed65f223cc3d83b3ec0550aa`, `make build-all`, `extra-assets-json: '[]'`, `go test ./...`, or `go vet ./...`.
* `TestMakefileVerify`: from the module root, GNU `make -n verify GO_SHARED=go-shared-under-test`. The output contains `python3 "go-shared-under-test/scripts/check_disclosure.py" --repo .`, `gofmt -l`, `go mod tidy -diff`, `go mod verify`, and `go run golang.org/x/vuln/cmd/govulncheck@v1.10.0 ./...`. The Makefile's `verify` dependency list is `mod-check fmt-check lint vet test vuln check-records markdownlint disclosure-check`. The `verify` recipe does not contain `go fmt`, `actionlint`, `check_ci_filters.py`, `build-all`, or `-mode=binary`.

After the tests compile, run `go mod tidy` again from the module root and require `go mod tidy -diff` to exit 0. The Phase 1 version assertions still hold. `mcplib` stays `v1.6.0` because the harness imports it.

Module root in tests is the parent of `cmd/`, found by walking up from the test's working directory until `go.mod` names this module. Do not embed a machine path.

Each new test is first run against a deliberately broken copy, outside the tree: a stamp case whose expected kind is flipped, a workflow copy that still has `bridge-release`, a workflow copy that contains `path: ${{ runner.temp }}/go-shared`, a makefile copy that still has `main.RawVersion`, a spec copy that omits `installer`, an identity command that writes with `fmt.Println` while `OriginalStdout` is a different pipe, a makefile copy whose `disclosure-check` recipe is `python3 scripts/check_disclosure.py` with no `GO_SHARED` and no `--repo`, and a makefile copy whose `verify` dependencies omit `vuln`. Each of those must fail. Record the failure text in the execution record. The tree itself is not the broken copy.

## Phase 6 - Verification

Run these after the last product edit. Capture the exit status before any filter. A non-zero status stops the phase.

Session host, this package:

```bash
gofmt -l cmd/mcp-server-duckduckgo/update.go cmd/mcp-server-duckduckgo/update_test.go cmd/mcp-server-duckduckgo/identity.go cmd/mcp-server-duckduckgo/main.go cmd/mcp-server-duckduckgo/root.go
go test -count=1 ./cmd/mcp-server-duckduckgo
SELFUPDATE_STAMP_CASE=release go test -count=1 -ldflags "-X github.com/maccavelli/go-selfupdate-lib/buildinfo.version=v9.9.9 -X github.com/maccavelli/go-selfupdate-lib/buildinfo.kind=release" ./cmd/mcp-server-duckduckgo -run TestStampCase
SELFUPDATE_STAMP_CASE=dirty go test -count=1 -ldflags "-X github.com/maccavelli/go-selfupdate-lib/buildinfo.version=v9.9.9-dirty -X github.com/maccavelli/go-selfupdate-lib/buildinfo.kind=release" ./cmd/mcp-server-duckduckgo -run TestStampCase
SELFUPDATE_STAMP_CASE=capital go test -count=1 -ldflags "-X github.com/maccavelli/go-selfupdate-lib/buildinfo.version=v9.9.9 -X github.com/maccavelli/go-selfupdate-lib/buildinfo.kind=Release" ./cmd/mcp-server-duckduckgo -run TestStampCase
SELFUPDATE_STAMP_CASE=local-tag go test -count=1 -ldflags "-X github.com/maccavelli/go-selfupdate-lib/buildinfo.version=v9.9.9 -X github.com/maccavelli/go-selfupdate-lib/buildinfo.kind=local" ./cmd/mcp-server-duckduckgo -run TestStampCase
```

`gofmt -l` prints nothing. Each stamp command's output contains `PASS` and does not contain `SKIP`. A dirty working tree can add a revision suffix to `Identity().String()` and still leaves `KindRelease` in place for the `release` case. A `SKIP` in that output is a failure. Deviation 2026-11-17: the four commands are run again with `-v`. Without `-v`, `go test` prints an `ok` line and does not print `PASS`. The `TestStampCase` assertions stay.

`make verify` includes `test` and `vuln`. `test` is `go test -v ./...`, the same suite the validate job runs today, with `go test`'s default timeout. This plan adds no skip. `test` is a heavy run on the session host. `vuln` and `markdownlint` need the network. An unreachable vulnerability database or registry fails the phase. Do not add a skip.

Before `make verify`, stage every path the Phase 7 commit will contain, including `identity.go` and the docs. `git ls-files` is the index, so an unstaged new file is invisible to `fmt-check` and `disclosure-check`. Staging is not the commit. The MADR frontmatter is `accepted`. This plan's frontmatter is `in-progress` from the start of execution. `check-records` accepts both.

Install the lint binary the CI workflow installs, then leave `GOBIN` unset:

```bash
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0
```

The Makefile default is `$(go env GOPATH)/bin/golangci-lint`. The install hint in the recipe now names `@v2.15.0`. Do not install `@latest`. Do not edit `.golangci.yml` to clear a finding. A new finding stops the phase and goes to the owner. `GO_SHARED` defaults to `../go-shared`, the sibling checkout whose `--self-test` passed in Phase 4. `run-heavy` runs on the session host and does not copy the tree. Omit `-j`.

```bash
~/.agents/skills/available-dev-hosts/scripts/run-heavy -- make verify
```

Search the changed files for a home directory, an account name, and a machine path. A hit is a stop. Describe the search in the execution record without quoting a hit.

`go version -m` on one locally built binary is optional and stays out of the commit. The stamp tests are the identity check. Do not also run a separate `go test ./...`, `go vet ./...`, `make lint`, `make check-records`, or `make markdownlint`. `verify` already runs them.

## Phase 7 - Records and the commit

After Phase 6's `make verify` has exited 0:

* Confirm the MADR frontmatter is `accepted` and this plan's frontmatter is `in-progress`. Stop when either word is something else. Do not flip the MADR again. Acceptance on 2026-11-17 already happened, and it did not edit 0003.
* Append this heading and paragraph at the end of `0003-MADR-update-dependency-pins.md`, leaving the deferred paragraph in place. The heading date is the day this step runs. Use 2026-11-17 when that is the day.

```markdown
### Amendment 2026-11-17 — self-update library

The separate decision in the last paragraph is
[0008-MADR-adopt-go-selfupdate-lib-v1.26.3.md](0008-MADR-adopt-go-selfupdate-lib-v1.26.3.md).
That paragraph stays.
```

* Set this plan's frontmatter `status` to `implemented` only when every completion criterion below has been met. A commit by itself does not make it `implemented`.
* Stage the pointer and the status edit. Run `make disclosure-check`, `make check-records`, and `make markdownlint`. All three exit 0. Do not run `make verify` a second time. The Go sources are unchanged since Phase 6.

This repository has no `scripts/go-precheck.sh` and no `make pre-add-check` target. Before `git commit`, `gofmt` every staged `.go` file. Run per-file `golint` on those files when the `golint` binary is on `PATH`. A `golint` finding stops the commit. If `golint` is absent, write that in the execution record and rely on `make lint`. Confirm `git rev-parse --path-format=absolute --git-path hooks` is the fleet global hooks directory or a wrapper that chains to it. Then `git commit --no-edit`. Do not pass a message flag, do not set `GIT_EDITOR=true`, and do not pass `--no-verify`. Do not push. Do not tag.

The commit is one commit because no earlier phase is green on its own. Phase 1's tidy drops the library when the import is absent. Phase 2 does not compile until Phase 5. Phase 5's workflow tests need Phases 3 and 4. Phase 4 can stop the run before any of that is commitable. Do not commit a red package to satisfy a per-phase commit. Do not push. Do not tag.

## Rollback

Before any tag, revert with new commits. Restore the `mcplib/selfupdate` imports in `update.go`, `update_test.go`, and `main.go`. Restore `version.go`, the `RawVersion` display assignment, and the four `-X main.RawVersion` / `-X main.RawBuildKind` pairs. Delete `identity.go` and `selfupdate-release.json`. Restore the hand-rolled `build` job, the workflow pin `8d4d89ff34ff79c1ed65f223cc3d83b3ec0550aa`, and `bridge-release: false`. Restore the `validate` steps `go test ./...`, `go vet ./...`, and the install step that runs `make lint`. Remove the `go-shared` fetch. Remove `GO_SHARED`, `fmt-check`, `mod-check`, `disclosure-check`, `vuln`, and `verify` from the Makefile. Restore the `lint` hint `@latest` only together with those targets, because that hint belongs to the tree this plan replaces. Drop the `go-selfupdate-lib` requirement with `go get` and `go mod tidy`, and confirm `mcplib` is still `v1.6.0` and the `go` and `toolchain` lines are unchanged. Drop the 0003 pointer. Do not rewrite history. A binary already installed is unchanged by the revert; only a later tag would publish either workflow. The generated scripts exist only as release assets, so a revert removes them from the next tag. The `go-shared` script stays. It belongs to that repository's 0003 plan.

## Completion Criteria

* `go list -m github.com/maccavelli/go-selfupdate-lib` is `v1.26.3`, and the Phase 1 version assertions all hold.
* `grep -R 'maccavelli/mcplib/selfupdate' --include='*.go' .` finds nothing, and `grep -R 'github.com/maccavelli/mcplib"' --include='*.go' .` still finds the harness files named in the grounding reference.
* `version.go` is gone. `main.RawVersion` and `main.RawBuildKind` occur nowhere in the `Makefile`.
* `update.go` contains `cli.Flags`, `Flags.Request`, `cli.Run`, `DefaultHTTPClient`, `DiscardReporter`, and `NonInteractiveConfirmer`, and it does not contain `os.Exit`, `cli.Exit`, `NewTextReporter`, or `NewTerminalConfirmer`.
* `identity.go` writes `buildinfo.Identity().String()` to `OriginalStdout` and carries `selfupdate.skip-config`.
* `selfupdate-release.json` parses, names `installer` `mcp-server-duckduckgo`, and `InstallerScripts` is `install.sh` then `install.ps1`.
* `.github/workflows/ci.yml` contains the peeled SHA `1dffdc9dee24a7534513dc51c9039408be37c09b` on both workflow calls, passes `needs.build.outputs.extra-assets-json`, and does not contain `bridge-release`, `make build-all`, or `path: ${{ runner.temp }}/go-shared`. The `validate` job still installs golangci-lint `v2.15.0`, fetches `https://github.com/maccavelli/go-shared.git` at the Phase 4 hash into `${{ runner.temp }}/go-shared`, and runs `make verify`.
* `make -n disclosure-check` runs `check_disclosure.py --repo .` through `"$(GO_SHARED)"`. `make -n vuln` runs `go run golang.org/x/vuln/cmd/govulncheck@v1.10.0 ./...`. `verify`'s dependency list is `mod-check fmt-check lint vet test vuln check-records markdownlint disclosure-check`.
* The package tests, the four stamp invocations, and `run-heavy -- make verify` exit 0.
* The eight planted broken copies failed before the real tests were counted, and the execution record quotes those failures.
* `make check-records` and `make markdownlint` exit 0 after the 0003 pointer and the status edits.
* No tag was created and no push was made by this plan.
* Frontmatter here is `implemented` only after the commit that contains the change. Until the owner says to proceed, it stays `proposed`.

## Execution record

The owner said to proceed on 2026-11-17. This plan's frontmatter `status` is `in-progress`. The MADR frontmatter is `accepted`. Phases 2, 1, and 3 are in the working tree and are uncommitted. Phase 4 stopped. Phases 5, 6, and 7 have not started. No tag was created and no push was made. The scratch modules cited by the MADR remain evidence for the decision.

### Amendment 2026-11-17 — install scripts

The owner asked for install scripts while this plan was still proposed. Phase 3 now calls `build-selfupdate-release.yml`, adds `selfupdate-release.json` with `installer`, and adds the `identity` command those scripts and the identity job run. The first draft only retargeted the publish `uses:` line and kept `extra-assets-json` as `[]`. That draft was not executed. The same day's MADR revision records the release workflow as part of the decision and leaves `windows/arm64` out (MADR E18). The phases above already match that shape.

### Amendment 2026-11-17 — disclosure check and verify

The owner asked to scope in the disclosure and verify checks and to house them in `go-shared`, while this plan was still proposed. Phase 4 now adds `make disclosure-check` and `make verify`, and it places `go-shared` on the runner. The acceptance review below is the fetch that replaced an `actions/checkout` path outside the workspace. The disclosure script is `go-shared` `0003-MADR-house-disclosure-check.md`. This plan does not write that script. The earlier revision had left both targets out of scope. No phase has run.

### Amendment 2026-11-17 — govulncheck

The owner asked for `govulncheck` in this repository while this plan was still proposed. `vuln` is now a dependency of `verify`, and its recipe is `go run golang.org/x/vuln/cmd/govulncheck@v1.10.0 ./...`. On this date that command exited 0 against the current module and printed `No vulnerabilities found.` `go.mod` and `go.sum` stayed unchanged. The earlier revision had left the scan out of `verify`. No phase has run.

### Amendment 2026-11-17 — acceptance review

The owner accepted the MADR on this date and asked for this plan to be checked against that record and the tree. The MADR status is `accepted`. This plan stays `proposed` until the owner says to proceed. No phase has run.

The review changed the execution shape. Phase 1's `go mod tidy` runs after the Phase 2 import, because tidy drops a requirement no file imports. Phases 2 and 5 are one compile unit: `newUpdater`'s type is `func(io.Writer)`, `newTestRoot` does not call `Execute`, and `TestBuildKindMapping` plus `TestVersionStampIsEffective` leave with `version.go`. The validate job fetches `go-shared` with git into `${{ runner.temp }}`. `actions/checkout` at `3d3c42e5aac5ba805825da76410c181273ba90b1` throws for a path outside `GITHUB_WORKSPACE` (MADR E21). `fmt-check` pipes `git ls-files -z` to `xargs -0` and records the command status, because GNU Make 3.81 on the session host does not apply `.SHELLFLAGS` and a shell variable keeps no bytes after the first NUL. `TestReleaseWorkflowPin` reads the workflow file and does not open this plan. Phase 6 uses `run-heavy` on the session host, installs golangci-lint `v2.15.0`, and stages the commit set before `make verify`. Phase 7 does not accept the MADR again. The 0003 pointer is still that phase. The one commit is after Phase 6, because no earlier boundary is green.

### Execution 2026-11-17 — stopped at Phase 4

`go-shared` HEAD is `d62e2a23b7228eb81af8b02170b985b38bbe4679`. Its `scripts/` directory contains `check_libs.py` and `check_records.py`. `scripts/check_disclosure.py` is absent. `python3 ../go-shared/scripts/check_disclosure.py --self-test` exited 2 because that path is not a file. `git -C ../go-shared cat-file -e d62e2a23b7228eb81af8b02170b985b38bbe4679:scripts/check_disclosure.py` exited 128 and reported that the path does not exist in that commit. No `GO_SHARED_SHA` was written. The public depth-1 fetch was not run, because no commit containing the script was identified. The sibling checkout was left as found: staged `0001-PLAN-scaffold-go-shared-workspace.md` and `0002-PLAN-shared-library-compatibility-check.md`, unstaged README files, and an untracked proposed 0003 pair. This repository did not vendor the script.

Phase 4 added no Makefile target and left the `validate` job on `go test ./...`, `go vet ./...`, and `make lint`. Phase 5 was not started.

### Execution 2026-11-17 — Phase 4 resumed, Phase 5 stopped

The owner said to resume. Sibling `go-shared` HEAD is `3e92bdae9c3e8fd6c662f891e803a1135ca8b513`, which is `origin/main`. `scripts/check_disclosure.py` is in that commit. `--self-test` exited 0. `git cat-file -e` of that path exited 0. A tokenless `git fetch --depth 1` of that commit from `https://github.com/maccavelli/go-shared.git` exited 0. The hash is 40 lowercase hexadecimal characters and is distinct from `1dffdc9dee24a7534513dc51c9039408be37c09b`, `10316bca82c00aeefde4bc614d183d343ed5818e`, and `8d4d89ff34ff79c1ed65f223cc3d83b3ec0550aa`.

Phase 4 wrote that hash as `GO_SHARED_SHA` in `.github/workflows/ci.yml` and replaced the validate job's `go test ./...`, `go vet ./...`, and `make lint` steps with the fetch and `make verify`. The Makefile gained `GO_SHARED ?= ../go-shared`, `fmt-check`, `mod-check`, `disclosure-check`, `vuln`, and `verify`. The lint install hint names `@v2.15.0`.

Phase 5 rewrote `update_test.go` and added `release_test.go`. `go test -count=1 ./cmd/mcp-server-duckduckgo` failed two tests. `TestUpdateUsesCallerContext` at `update_test.go` returned `selfupdate: target <the go test binary> is outside allowed roots`. `errors.Is` against `context.Canceled` was false. The constructor had already returned. `TestCheckAndJSONStreams` returned the same roots error on `--check`, before `ErrUpdateAvailable`. `StandaloneInstaller.ResolveTarget` ignores the context argument and refuses an executable outside the home directory. A `go test` binary is in the system temporary directory. Empty `InstallOptions` and `defaultNewUpdateUpdater` were left as the plan wrote them. The assertions were not loosened. The eight planted copies were not run. Phase 6 has not started. No commit, push, or tag.

### Deviation 2026-11-17 — re-exec under the home directory

The owner chose re-exec under the home directory on 2026-11-17, after `go test -count=1 ./cmd/mcp-server-duckduckgo` failed `TestUpdateUsesCallerContext` and `TestCheckAndJSONStreams`.

What was found. Both tests returned an error whose text is `selfupdate: target <the go test binary> is outside allowed roots`. `errors.Is` against `context.Canceled` was false. `--check` had not returned `ErrUpdateAvailable`. `selfupdate.New` had already returned. At peeled `1dffdc9dee24a7534513dc51c9039408be37c09b`, `Updater.Run` reports `EventResolvingTarget`, then calls `Installer.ResolveTarget`, then discovers the release (`selfupdate/updater.go`). `StandaloneInstaller.ResolveTarget` ignores the context and calls `resolveTarget` (`selfupdate/standalone.go`). `allowedRoots` for an empty `TargetPolicy.AllowedRoots` is the home directory (`selfupdate/target.go`). A `go test` binary is placed in the system temporary directory. That directory is outside the home directory. `NewGitHubSource` still does not dial while the constructor runs (MADR E10).

What changes. `TestUpdateUsesCallerContext` and `TestCheckAndJSONStreams` copy this process's executable into a new directory under the home directory and run that copy with `-test.run` limited to the one test. The parent removes the directory when it returns. When the executable is already under the home directory, the test body runs in the current process. `defaultNewUpdateUpdater` stays the cancelled-context constructor. The fake updater still passes `NewStandaloneInstaller` an empty `InstallOptions`. `TestUpdateUsesCallerContext` still requires `errors.Is(err, context.Canceled)`. `TestCheckAndJSONStreams` still requires `ErrUpdateAvailable`, an empty command-output buffer, the phrases `mcp-server-duckduckgo: update available:` and `local build: apply requires --force`, and a protocol buffer containing `"kind":"result"` and `"exit_code":10`. The file in scope for this deviation is `cmd/mcp-server-duckduckgo/update_test.go`.

The Phase 5 sentences for those two tests stay, with the deviation annotation on each bullet.

### Execution 2026-11-17 — Phase 5 re-exec

`gofmt -l` on `main.go`, `root.go`, `update.go`, `identity.go`, `update_test.go`, and `release_test.go` printed nothing. `go test -count=1 ./cmd/mcp-server-duckduckgo` exited 0. `TestUpdateUsesCallerContext` and `TestCheckAndJSONStreams` re-executed from a copy under the home directory. The parent removed that directory. No re-exec directory remained afterward.

Eight copies outside the tree each exited 1.

* A stamp case whose expected kind was `KindLocal` for `SELFUPDATE_STAMP_CASE=release`, built with version `v9.9.9` and kind `release`, failed at `release_test.go:40` with `release stamp: kind release current "v9.9.9"`.
* A workflow copy containing `bridge-release` failed at `release_test.go:255` with `workflow contains bridge-release`.
* A workflow copy containing `path: ${{ runner.temp }}/go-shared` failed at `release_test.go:255` with `workflow contains path: ${{ runner.temp }}/go-shared`.
* A Makefile copy whose local `build` recipe also passed `-X main.RawVersion=$(VERSION)` failed at `release_test.go:65` with `old stamp names remain`.
* A spec copy that omitted `installer` failed at `release_test.go:199` with `scripts = []`.
* An identity command that called `fmt.Println` while `OriginalStdout` was a different pipe failed at `release_test.go:132` with `identity wrote "", want "dev (local)\n"`.
* A Makefile copy whose `disclosure-check` recipe was `python3 scripts/check_disclosure.py` failed at `release_test.go:287` with `make -n verify missing python3 "go-shared-under-test/scripts/check_disclosure.py" --repo .`.
* A Makefile copy whose `verify` dependencies omitted `vuln` failed at `release_test.go:271` with `verify line = "verify: mod-check fmt-check lint vet test check-records markdownlint disclosure-check ## Runs the local and CI quality checks"`.

`go mod tidy` exited 0. `go mod tidy -diff` exited 0. `go.mod` still has one `go 1.26.11` line and one `toolchain go1.27.4` line. `go list -m` printed `github.com/maccavelli/go-selfupdate-lib v1.26.3`, `github.com/maccavelli/mcplib v1.6.0`, `golang.org/x/sys v0.50.0`, `golang.org/x/mod v0.43.0`, `golang.org/x/term v0.48.0`, `github.com/ysmood/fetchup v0.3.0`, and `github.com/mxschmitt/playwright-go v0.6401.0`.

Phase 6 has not started. No commit, push, or tag.

### Execution 2026-11-17 — Phase 6 stamp output

`gofmt -l` on `update.go`, `update_test.go`, `identity.go`, `main.go`, `root.go`, and `release_test.go` printed nothing and exited 0. `go test -count=1 ./cmd/mcp-server-duckduckgo` exited 0. The four `SELFUPDATE_STAMP_CASE` commands in this phase each exited 0. Each printed one line that begins with `ok` and names this module. That line does not contain `PASS` or `SKIP`. `go test` prints `PASS` when `-v` is set. These commands do not pass `-v`. `make verify` has not started. No commit, push, or tag. The phase text asks for `PASS` in that output, so the phase is stopped on that check.

### Deviation 2026-11-17 — stamp commands use -v

The Phase 6 question was declined. The four commands had exited 0 and each had printed one `ok` line. That line contains neither `PASS` nor `SKIP`. `go test` prints `PASS` when `-v` is set. The same four commands are run again with `-v`. `TestStampCase` stays as written. A `SKIP` in the verbose output still fails the check.

Each of the four `-v` commands exited 0. Each output contains `--- PASS: TestStampCase (0.00s)` and `PASS`, and does not contain `SKIP`.

### Execution 2026-11-17 — Phase 6 verify stopped at disclosure-check

The Phase 7 file set was staged. `GOBIN` and `GOLANGCI_LINT` were unset. `go install` of `golangci-lint@v2.15.0` exited 0, and `golangci-lint version` reported `2.15.0`. Sibling `go-shared` HEAD is `3e92bdae9c3e8fd6c662f891e803a1135ca8b513`, and `scripts/check_disclosure.py` is present. A scan of the staged file set for a macOS home prefix, a Linux home prefix, the account name, and a tab in a Markdown file found 0 hits.

`run-heavy -- make verify` exited 2. `go mod tidy -diff` and `go mod verify` exited 0. `golangci-lint` reported `0 issues.` `go vet ./...` exited 0. `go test -v ./...` exited 0, including `TestUpdateUsesCallerContext`, `TestCheckAndJSONStreams`, and a skip of `TestStampCase` because `SELFUPDATE_STAMP_CASE` was empty. `govulncheck` printed `No vulnerabilities found.` `check-records` exited 0. `markdownlint-cli2` v0.23.2 reported `0 issues` in 31 files. `disclosure-check` then exited 1. Its summary was `disclosure: 104 files scanned, 17 finding(s), 0 unreadable`. All 17 findings are in the repository root `README.md`, which is unchanged from `HEAD` and was not part of the staged set. They are example install paths: a Linux home prefix, a macOS home prefix, and one Windows home prefix. The account name is absent. The segments are not angle-bracket placeholders. No commit, push, or tag.

### Deviation 2026-11-17 — example paths in the root README

The disclosure question was declined. `disclosure-check` reported 17 findings, all in the root `README.md`, which matches `HEAD`. The segments are the example words `youruser` and `YourUser` after a Linux home prefix, a macOS home prefix, or a Windows home prefix. The account name is absent. In `go-shared` `scripts/check_disclosure.py`, `is_placeholder` accepts an empty segment, `~`, a segment that starts with `$`, or a segment wrapped in `%`. A line that carries a full angle-bracket token is accepted before that function runs. The example segments become `<user>`. The disclosure script stays. The lines stay install examples. `README.md` is added to the staged set. `make verify` runs again.

### Execution 2026-11-17 — Phase 6 verify exited 0

The second `run-heavy -- make verify` exited 0. `go mod tidy -diff` and `go mod verify` exited 0. `golangci-lint` reported `0 issues.` `go vet ./...` exited 0. `go test -v ./...` exited 0. `TestStampCase` skipped because `SELFUPDATE_STAMP_CASE` was empty, which is the test's own empty-variable skip. `govulncheck` printed `No vulnerabilities found.` `check-records` exited 0. `markdownlint-cli2` v0.23.2 reported `0 issues` in 31 files. `disclosure-check` printed `disclosure: 104 files scanned, 0 finding(s), 0 unreadable`. The plan frontmatter stays `in-progress`. The MADR frontmatter stays `accepted`. No commit, push, or tag.

### Execution 2026-11-17 — commit authorized

The owner said to commit it all to main and push. `gofmt -l` on the staged Go files printed nothing. `golint` on those files printed nothing. `go list -m` still prints `github.com/maccavelli/go-selfupdate-lib v1.26.3` and `github.com/maccavelli/mcplib v1.6.0`. `go 1.26.11` and `toolchain go1.27.4` stay. No Go file imports `mcplib/selfupdate`. The harness files still import `mcplib`. `version.go` is gone. The Makefile has neither `main.RawVersion` nor `main.RawBuildKind`. `update.go` calls `flags.Request`, `cli.Run`, `DefaultHTTPClient`, `DiscardReporter`, and `NonInteractiveConfirmer`. This frontmatter `status` is `implemented` in the commit. The MADR frontmatter stays `accepted`.


Phase 2 ran before the module commands. `version.go` was deleted. `update.go` builds the updater in `defaultNewUpdateUpdater`, binds `cli.Flags`, and returns the error from `cli.Run`. `identity.go` writes `buildinfo.Identity().String()` to `OriginalStdout`. `root.go` saves `protocolStdout`, sets `RootCmd.Version` from `buildinfo.Identity().Current()` with one leading `v` removed, and registers `newIdentityCmd`. `main.go` imports `github.com/maccavelli/go-selfupdate-lib/selfupdate`. The four Makefile recipes pass `-X github.com/maccavelli/go-selfupdate-lib/buildinfo.version` and `-X github.com/maccavelli/go-selfupdate-lib/buildinfo.kind`. `gofmt -l` on `main.go`, `root.go`, `update.go`, and `identity.go` printed nothing. `go build` of `./cmd/mcp-server-duckduckgo` exited 0. `go test` was not run. `update_test.go` was not edited. It still imports `github.com/maccavelli/mcplib/selfupdate` and still names `newUpdater`, `RawVersion`, and `RawBuildKind`, including `TestBuildKindMapping`.

Phase 1 ran after that import existed. `go get github.com/maccavelli/go-selfupdate-lib@v1.26.3` exited 0 and added `v1.26.3`. `go mod tidy` exited 0. `go.mod` still has one `go 1.26.11` line and one `toolchain go1.27.4` line. `go list -m` printed `github.com/maccavelli/go-selfupdate-lib v1.26.3`, `github.com/maccavelli/mcplib v1.6.0`, `golang.org/x/sys v0.50.0`, `golang.org/x/mod v0.43.0`, `golang.org/x/term v0.48.0`, `github.com/ysmood/fetchup v0.3.0`, and `github.com/mxschmitt/playwright-go v0.6401.0`. `go mod tidy -diff` exited 0. `go.mod` gained that one require line. `go.sum` gained the two sum lines for `v1.26.3`. Both modules stay because `update_test.go` still imports `mcplib/selfupdate`.

Phase 3 wrote `selfupdate-release.json`: schema 1, product `mcp-server-duckduckgo`, package `./cmd/mcp-server-duckduckgo`, identity args `identity`, platforms linux/amd64, darwin/arm64, and windows/amd64, packaging `binary`, installer name `mcp-server-duckduckgo`, and no hooks, extras, prerelease channels, or tags. `.github/workflows/ci.yml` calls `build-selfupdate-release.yml` and `publish-selfupdate-release.yml` at peeled `1dffdc9dee24a7534513dc51c9039408be37c09b`. The release job runs when `needs.build.outputs.rehearsal == 'false'` and passes `artifact-name`, `products-json`, `platforms-json`, `extra-assets-json`, and `prerelease-channels-json`. The file has no `bridge-release`, no `make build-all`, no mcplib pin `8d4d89ff34ff79c1ed65f223cc3d83b3ec0550aa`, and no `module-dir`, `extras-artifact-name`, `retention-days`, or `sign` input. The `validate` job still checks out `actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1`, uses `actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e`, and installs golangci-lint `v2.15.0` before `make lint`.

The accepted MADR and the docs index rows for 0008 stay the uncommitted acceptance rewrite on top of `1335d2f7f144e30567f233a4b667bba5216e99d5`. Resume at Phase 4 after `go-shared` 0003 has landed `scripts/check_disclosure.py`, `python3 ../go-shared/scripts/check_disclosure.py --self-test` exits 0, `git cat-file -e` finds that path in the sibling HEAD, and a tokenless `git fetch --depth 1` of that 40-character commit from `https://github.com/maccavelli/go-shared.git` succeeds. That hash has to be distinct from `1dffdc9dee24a7534513dc51c9039408be37c09b`, `10316bca82c00aeefde4bc614d183d343ed5818e`, and `8d4d89ff34ff79c1ed65f223cc3d83b3ec0550aa`.
