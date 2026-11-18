# Updating

Task index: [docs/README.md](../README.md).

`update` checks GitHub releases for `maccavelli/mcp-server-duckduckgo`,
verifies the checksum, and can replace this executable. It creates no
cache directory and no `config.yaml`. It starts no datastore, browser,
search engine, tool registry, transport, or MCP server.

```bash
mcp-server-duckduckgo update
mcp-server-duckduckgo update --check
mcp-server-duckduckgo update --version v1.2.0
```

## Flags

`cmd/mcp-server-duckduckgo/update_test.go` `TestUpdateFlagSurface` requires these flags, and `-y` as the shorthand for `--yes`:

| Flag | What it does |
| :--- | :--- |
| `--check` | Report whether an update is available. Install nothing. |
| `--yes`, `-y` | Install without asking. |
| `--force` | Replace a local build, or reinstall the same version. |
| `--dry-run` | Download and verify the release. Install nothing. |
| `--json` | Write JSON Lines to the preserved stdout. Everything else goes to stderr. |
| `--version vX.Y.Z` | Install that exact release. |
| `--channel NAME` | Passed through on the request. This program uses `NewStrictVersionPolicy`, and `selfupdate-release.json` lists no prerelease channels. |

`TestUpdateRejectsContradictoryFlags` runs `--check` together with
`--yes`, `--force`, and `--dry-run`. Each pair fails before the updater
is constructed. `cli.Flags.Request` returns these errors:

- `selfupdate: --check and --yes are contradictory`
- `selfupdate: --check and --force are contradictory`
- `selfupdate: --check and --dry-run are contradictory`

A positional argument is rejected.

## Exit codes

| Code | Meaning |
| :--- | :--- |
| 0 | Up to date, declined, or installed. |
| 10 | `--check` found an actionable target. |
| 1 | Any other error. |

`main` prints an error to stderr when the code is not 10. Exit 10 stays quiet on stderr so a check can be scripted.

## Local builds and output

A binary built with the default `local` kind is refused unless `--force` is set. `make build` stamps that kind. The text of a local `--check` includes `local build: apply requires --force`.

Text from `update` is written to stderr. With `--json`, events and the
final result object are JSON Lines on the preserved stdout. That stream
is the process stdout from before the server aliases it to stderr.
The result line is the last object and carries `exit_code`.

`GH_TOKEN` or `GITHUB_TOKEN`, when set, is sent only to the GitHub API origin to raise the rate limit.

If the binary was installed by a package manager, update it with that manager. `update` replaces the executable it is running.
