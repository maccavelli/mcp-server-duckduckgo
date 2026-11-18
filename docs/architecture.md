# Architecture

How this repository is put together, as it is now. The task index and the decision records are in [README.md](README.md).

## What it is

This repository is the source of truth for the server.

`cmd/mcp-server-duckduckgo` is a stdio MCP server. With no subcommand,
`root.go` runs `serve`. `Execute` saves the process stdout, then points
`os.Stdout` at stderr, so logs do not share the JSON-RPC stream.
Protocol bytes go to the saved stdout.

`serve` loads configuration, opens the scraping store, registers tools,
and serves MCP on stdio until interrupt, `SIGTERM`, or `SIGHUP`.
When `MCP_ORCHESTRATOR_OWNED` is not `true`, it holds `run.lock` in the
cache directory so a second standalone process stops.

## Commands

| Command | Role |
| :--- | :--- |
| `serve` | The default. MCP over stdio. |
| `update` | GitHub release check and optional replacement of this executable. Skips configuration. |
| `identity` | Prints the build identity to the preserved stdout. Skips configuration. The installers run it. |

## Cache and store

The cache directory is `os.UserCacheDir()` joined with
`mcp-server-duckduckgo`. `config.Load` creates `config.yaml` there when
the file is absent. The template is comments. No production caller reads
keys from it.

`serve` opens `scraping.db` in that directory through `internal/store`.
The file holds VQD tokens, cookies, user-agent rotation, and provider
health. A failure to open the file uses an in-memory store for that process.

## Tools

`internal/handler/search` registers `search_web`, `search_news`,
`search_books`, and `fetch_url`. `internal/handler/media` registers
`search_images` and `search_videos`. `internal/handler/system` registers
`get_internal_logs`.

Web, news, image, and video search live in `internal/engine`.
`search_books` calls `https://openlibrary.org/search.json`.
`fetch_url` and `fetch_content` use the HTTP fetcher in
`internal/engine/fetch.go`. A 403 or a challenge page tries Playwright
once the driver install has been attempted.

## Release

`selfupdate-release.json` names one binary product and three platforms:
`linux/amd64`, `darwin/arm64`, and `windows/amd64`.
`.github/workflows/ci.yml` runs `make verify` in `validate`, then the
`go-selfupdate-lib` `v1.26.3` build workflow. A `v*` tag publishes.
Any other ref rehearses, and the publish job is skipped. The sign job
is skipped because the workflow does not set signing.

`make verify` is `mod-check`, `fmt-check`, `lint`, `vet`, `test`, `vuln`,
`check-records`, `markdownlint`, and `disclosure-check`. CI and a local
checkout run that same target. `disclosure-check` calls
`scripts/check_disclosure.py` from the `go-shared` checkout in `GO_SHARED`.

## Tree

```text
README.md                 product entry; links to docs/README.md
CLAUDE.md                 project instructions for agents
AGENTS.md                 symlink to CLAUDE.md
docs/
  README.md               record index and the "I want to…" matrix
  architecture.md         this file
  decisions/              MADR and PLAN records
  guides/                 install, update, client, tools, configuration, development, release
selfupdate-release.json   product, platforms, installer name
.github/workflows/ci.yml  validate, build, publish
```

`docs/reports/` is not in the tree. The first report creates it.

## What is not here

- Secrets.
- The operator's git hooks and skill checkouts. Those stay on the machine that clones the repository.
- A `windows/arm64` release asset.
- A second documentation tree. Search, fetch, and the book catalog are this repository's product.
