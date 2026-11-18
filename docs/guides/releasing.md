# Releasing

Task index: [docs/README.md](../README.md).

Release facts below were checked against tag `v1.2.0`, which points at commit `5cfff6129cabab78f36f093a22e3bb5c1ce504fa`.

## Spec

`selfupdate-release.json` is the release spec:

- One product, `mcp-server-duckduckgo`, package `./cmd/mcp-server-duckduckgo`, `identity_args` of `identity`.
- Packaging `binary`.
- Platforms `linux/amd64`, `darwin/arm64`, and `windows/amd64`.
- Installer name `mcp-server-duckduckgo`. No hooks and no `env_prefix`.

## Workflow

`.github/workflows/ci.yml` runs on a push to `main`, on a tag matching `v*`, and on a pull request.

1. `validate` checks out the repository, sets up Go from `go.mod`,
   downloads modules, fetches `go-shared` at
   `3e92bdae9c3e8fd6c662f891e803a1135ca8b513`, installs golangci-lint
   `v2.15.0`, and runs `make verify`.
2. `Build release` calls `build-selfupdate-release.yml` from `go-selfupdate-lib` `v1.26.3` at peeled commit `1dffdc9dee24a7534513dc51c9039408be37c09b`, with `spec-path: selfupdate-release.json`.
3. `Publish self-update release` calls `publish-selfupdate-release.yml` at that same commit when `needs.build.outputs.rehearsal` is `false`.

A push to `main` rehearses: the binaries are built and the identity jobs run, and publish is skipped. A `v*` tag stamps the tag as a release build and publishes.

Signing inputs are unset. The sign job stays skipped.

## Assets

A published release contains:

- `mcp-server-duckduckgo-linux-amd64`
- `mcp-server-duckduckgo-darwin-arm64`
- `mcp-server-duckduckgo-windows-amd64.exe`
- `SHA256SUMS`
- `install.sh`
- `install.ps1`

`v1.2.0` published that set. There is no `windows/arm64` asset. Install and update behavior for those files is in [Installing](installing.md) and [Updating](updating.md).
