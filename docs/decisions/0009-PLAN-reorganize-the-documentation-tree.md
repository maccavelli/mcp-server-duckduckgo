---
status: implemented
date: 2026-11-18
associated-madr: "0009-MADR-reorganize-the-documentation-tree.md"
---
# Implement the short README and the guide library

Associated MADR: [0009-MADR-reorganize-the-documentation-tree.md](0009-MADR-reorganize-the-documentation-tree.md)

## Goal

* The root `README.md` is a short overview plus the two latest-release installer one-liners, a table of contents, an "I want to…" table, and the three `update` command lines.
* `docs/guides/` contains the seven guides named in the MADR, and each procedure matches the sources named in that record.
* `docs/README.md` indexes `0009` and its "I want to…" rows start from a task.
* `docs/architecture.md` describes the current process and carries no history.
* No product source, workflow, module file, or release spec changes.

## Scope

In scope:

* `README.md`
* `docs/README.md`
* `docs/architecture.md`
* `docs/guides/installing.md`
* `docs/guides/updating.md`
* `docs/guides/connecting-a-client.md`
* `docs/guides/mcp-tools.md`
* `docs/guides/configuration.md`
* `docs/guides/developing.md`
* `docs/guides/releasing.md`

Out of scope:

* Records `0001` through `0008`, except the hand-written index rows this plan adds for `0009`.
* `go.mod`, `go.sum`, `Makefile`, `.github/workflows/ci.yml`, `selfupdate-release.json`, and every `.go` file.
* `CHANGELOG.md`, `SECURITY.md`, `CONTRIBUTING.md`, `CODEOWNERS`, and `docs/reports/`.
* A tag, a commit, and a push. Those wait for a later instruction.
* IDE configuration paths copied from the current README.

## Grounding Reference

The decision and evidence E1 through E5 are in [0009-MADR-reorganize-the-documentation-tree.md](0009-MADR-reorganize-the-documentation-tree.md). `documentation-writing` is the layout. `scripts/check_records.py` checks links in the root README, `docs/README.md`, `docs/architecture.md`, and `docs/guides/`, and it does not generate the index. PLAN status in this repository is `proposed` until execution starts. `implemented` is the status after the acceptance criteria pass. `complete` is rejected by `check_records.py`.

## Prerequisites

* `HEAD` is `5cfff6129cabab78f36f093a22e3bb5c1ce504fa` on `main`, and `v1.2.0` points at that commit.
* The owner has accepted the MADR and said to execute this plan.
* Facts below are copied from the named source. A sentence whose source is not in this plan or the MADR is not added during execution.
* Placeholders in examples use a full `<user>` token where a home segment must be shown. `~`, `%LOCALAPPDATA%`, and `$HOME` need no account segment.
* `make verify` is a heavy run. This plan does not run it. The gates in Phase 4 are the docs gates.

## Phase 1 - Root README

Replace `README.md`. ~~Keep the existing mirror-notice blockquote as the first block.~~ The first paragraph states that this repository is the source of truth. Then:

1. One H1, `MagicDuck MCP server`, and one paragraph: a stdio MCP server for web, news, image, and video search, Open Library book lookup, and page fetch. Name the current release `v1.2.0` and the platforms `linux/amd64`, `darwin/arm64`, and `windows/amd64`. Link [docs/README.md](../README.md).
2. An `## Install` section with exactly these two commands, and one sentence that the installer checks `SHA256SUMS`, installs to `~/.local/bin` or `%LOCALAPPDATA%\Programs\mcp-server-duckduckgo`, and runs `identity`. Link `docs/guides/installing.md`.

   ```bash
   curl -fsSL https://github.com/maccavelli/mcp-server-duckduckgo/releases/latest/download/install.sh | sh
   ```

   ```powershell
   irm https://github.com/maccavelli/mcp-server-duckduckgo/releases/latest/download/install.ps1 | iex
   ```

3. An `## Overview` section of short bullets for what the server does. Include `search_books` against Open Library. Do not name `search_suggest`, `read_url`, `MCP_ENDPOINT_API_PORT`, or `MCP_LOG_LEVEL`.
4. A `## Table of contents` list of the H2 headings in this file.
5. An `## I want to…` table whose rows are tasks. Each row links a guide or `docs/README.md`. The tasks are: install, update, connect a client, call a tool, set cache and environment, build and test, cut a release, and find a decision.
6. An `## Update` section with the three command lines now at the top of `README.md`, with the example tag written as `v1.2.0`, and a link to `docs/guides/updating.md`.
7. An `## Quick start` section before `## Install`. It contains the same two commands as item 2, then tells the reader to run `mcp-server-duckduckgo` with no arguments. MagicTools launches that binary in production. A standalone client uses the same command. Link `docs/guides/installing.md` and `docs/guides/connecting-a-client.md`. The table of contents includes this heading.

Remove the manual `mv` install, the IDE catalogue, the environment table, the tool table, the storage table, the emoji headings, and the `markdownlint-disable` comment. The file has to lint under the human profile with no disable comment.

## Phase 2 - Guides

Create `docs/guides/` by adding the seven files. Each file starts with an H1 and links back to `docs/README.md`. Required content:

`installing.md`

* The two one-liners from Phase 1. A tag URL installs that tag, so the pinned Unix command is `curl -fsSL https://github.com/maccavelli/mcp-server-duckduckgo/releases/download/v1.2.0/install.sh | sh`. The pinned Windows command that can take `-Version` is the scriptblock form from go-selfupdate-lib building-releases section 12, with this repository's URL. Do not pass `--version` to an installer that was already downloaded from that tag's URL.
* Default directories: `~/.local/bin` and `%LOCALAPPDATA%\Programs\mcp-server-duckduckgo`.
* Options that the library section documents and this spec can use: `--dir` / `-InstallDir`, `--version` / `-Version`, `--dry-run` / `-DryRun`, `--uninstall` / `-Uninstall`, `--allow-root`, `--verify-attestation` / `-VerifyAttestation`. Variable prefix `MCP_SERVER_DUCKDUCKGO` for `MCP_SERVER_DUCKDUCKGO_VERSION` and `MCP_SERVER_DUCKDUCKGO_INSTALL_DIR`.
* The installer downloads `SHA256SUMS` and the platform asset, checks the checksum, swaps the binary, leaves the previous as `mcp-server-duckduckgo.prev` (`.exe.prev` on Windows), and runs `identity`. No hooks are configured.
* Supported platforms are the three in `update.go`. `install.ps1` reads the machine architecture from the registry, so a Windows ARM64 machine is refused because this release has no `windows/arm64` asset.
* `irm … | iex` cannot pass parameters. Say that next to the one-liner.
* Exit codes `0`, `1`, `2`, and `3` as that library section states them.

`updating.md`

* Commands from `update.go`'s `Long` text: `update`, `update --check`, `update --version v1.2.0`. Exit `0` up to date or declined, `10` when `--check` finds a target, `1` on error.
* `--check` contradicts `--yes`, `--force`, and `--dry-run` before the updater is built. State that from the current `update` tests rather than from memory of an older flag set. Read `cmd/mcp-server-duckduckgo/update_test.go` while writing the section and name only flags that file or `cli.Flags` asserts.
* `update` and `identity` do not create the cache directory or `config.yaml`.
* A local build is refused without `--force`. `GH_TOKEN` or `GITHUB_TOKEN` is sent only to the GitHub API origin.
* Output stays off the JSON-RPC stdout. JSON mode is `--json`, and its result line goes to the preserved stdout.

`connecting-a-client.md`

* Production: the client connects to MagicTools. This binary is the downstream server. Do not invent a MagicTools config file.
* Standalone: the client launches `mcp-server-duckduckgo` with no arguments. `root.go` runs `serve`. Transport is stdio. Protocol stdout is the saved original stdout.
* When `MCP_ORCHESTRATOR_OWNED=true`, `serve` does not take the singleton lock (`mcplib.IsOrchestratorOwned`, `serve.go`).
* No table of IDE configuration paths.

`mcp-tools.md`

* One section per tool, with the name, the description string from the registration, and the input fields from the struct or schema.
* `search_web`, `search_news`, `search_books`: `query`, `max_results` default 5, `format` of `hybrid`, `json`, or `markdown` (default `json`, or `DDG_DEFAULT_FORMAT` when set), `fetch_content` default false. Source: `internal/handler/search/tools.go`.
* `search_books` calls Open Library `search.json`. A non-positive `max_results` becomes 5. Values above 50 become 50. An empty query does not call the network. Source: `internal/engine/books.go`.
* `search_images` and `search_videos`: `query`, `max_results` default 5, `format` as above. No `fetch_content`. Source: `internal/handler/media/tools.go`.
* `fetch_url`: required `url`. It is the page-fetch tool. `read_url` is not a tool.
* `get_internal_logs`: optional `session_id` and `max_lines`. Omitted or non-positive `max_lines` returns 25. The cap is 1000. Source: `mcplib` `v1.6.0` `diagnostic.go`, called from `internal/handler/system/tools.go` with domain `duckduckgo`.
* `search_suggest` is not a tool.

`configuration.md`

* Cache directory: `os.UserCacheDir()` joined with `mcp-server-duckduckgo`. Name the three platform results of `UserCacheDir`: `$XDG_CACHE_HOME` or `$HOME/.cache` on Linux, `$HOME/Library/Caches` on macOS, `%LocalAppData%` on Windows. Files: `config.yaml` (created with the comment template when absent), `run.lock` when the process is not orchestrator-owned, `scraping.db`.
* `config.yaml` has no keys the server reads. `config.GetString` is only called from the config test.
* Environment that changes behavior: `DDG_DEFAULT_FORMAT`, `OPENLIBRARY_CONTACT` (appended to the Open Library user agent when non-empty), `GH_TOKEN` and `GITHUB_TOKEN` for `update`, `MCP_ORCHESTRATOR_OWNED=true` to skip the lock, and `GOMEMLIMIT` / `GOMAXPROCS` which `serve` sets to `1024MiB` and `2` only when unset.
* `fetch` falls back to Playwright after a CAPTCHA or HTTP 403 (`internal/engine/fetch.go`). `playwright.Install` runs on that path. Do not tell the reader to install a browser before the first search.

`developing.md`

* Module floor `go 1.26.11` and `toolchain go1.27.4`, from `go.mod`.
* `make build` writes `dist/mcp-server-duckduckgo-<goos>-<goarch>` with `buildinfo.kind` defaulting to `local`. `make test` is `go test -v ./...`. `make verify` runs `mod-check`, `fmt-check`, `lint`, `vet`, `test`, `vuln`, `check-records`, `markdownlint`, and `disclosure-check`. Lint installs `golangci-lint` `v2.15.0`. `vuln` runs `govulncheck` `v1.10.0`. Disclosure uses the sibling `go-shared` checkout named by `GO_SHARED` (default `../go-shared`).
* `make fmt` rewrites files. `make fmt-check` does not.
* Layout: `cmd/mcp-server-duckduckgo` for the binary, `internal/engine` for search and fetch, `internal/handler` for tools, `internal/store` for BuntDB. Point at `docs/architecture.md` for the map.
* Commands a developer runs: `serve` (also the default), `update`, `identity`. `identity` prints `buildinfo.Identity().String()` and a newline to the preserved stdout.

`releasing.md`

* `selfupdate-release.json` lists one product, `package` `./cmd/mcp-server-duckduckgo`, `identity_args` `identity`, packaging `binary`, and the three platforms.
* A push to `main` rehearses. A `v*` tag publishes. The workflow calls go-selfupdate-lib `v1.26.3` at peeled commit `1dffdc9dee24a7534513dc51c9039408be37c09b`.
* Published assets are the three binaries, `SHA256SUMS`, `install.sh`, and `install.ps1`. Signing is unset, so the sign job stays skipped.
* `v1.2.0` is the release this plan's facts were checked against. Do not invent a newer tag.

## Phase 3 - Index, matrix, and architecture

* `docs/README.md` already gains the `0009` index rows and one "know why" row when this pair is added. During execution, rewrite `## I want to…` so the first rows are the Phase 1 tasks and point at the guides. Keep a "know why" row for each accepted decision, phrased as the reader's question, one row per decision rather than one row per paragraph. Drop rows that only repeat a plan's implementation steps.
* Rewrite `docs/architecture.md` as the current system. Include the command tree, the stdio serve path, the cache directory, the tool packages, Open Library, the Playwright fallback, the three release platforms, and the docs tree with `guides/` present and `reports/` still absent. State that CI runs `.github/workflows/ci.yml` and `make verify`. Delete the sentence that says CI is absent and the gates run only by hand. Keep history and rationale out. Decision links belong in `docs/README.md`.
* The root `README.md` "I want to…" table and the `docs/README.md` table must name the same guides for the same tasks.

## Phase 4 - Gates

Run these after the last edit. Capture each exit status before any filter.

1. `python3 scripts/check_records.py --check-all`
2. `npx --yes markdownlint-cli2@0.23.2`
3. `python3 ../go-shared/scripts/check_disclosure.py --repo .` when `../go-shared` is the checkout CI fetches (`3e92bdae9c3e8fd6c662f891e803a1135ca8b513` or a descendant that still contains that script). A missing sibling is reported, not replaced with a copied script.

Before calling step 1 green, copy the tree to a scratch directory outside this repository, break one guide link, and run `--check-all` there. That run must exit 1 and print `broken relative link:`. Delete the scratch copy. Do not break the real tree.

A content check, also outside the tree, asserts the finished root `README.md` contains both one-liners and the string `search_books`, and that it does not contain `search_suggest`, `read_url`, `MCP_ENDPOINT_API_PORT`, `MCP_LOG_LEVEL`, `Programs\duckduckgo`, or `/usr/local/bin`. The check must fail on a copy that puts `read_url` back into the overview.

## Rollback

Restore `README.md`, `docs/README.md`, and `docs/architecture.md` from `HEAD`, and remove `docs/guides/` if this plan created it. No migrated data. The `0009` pair stays; withdrawing the decision is a status change, not a file delete.

## Completion Criteria

* The seven guides exist and a reader can install, update, connect, call each tool, find the cache, build, and see what a tag publishes without opening a decision record.
* The root README matches Phase 1.
* `docs/architecture.md` matches the current command and CI, with no "we chose" paragraph.
* Phase 4 exits 0 on the real tree, and the two planted copies exited 1 first.
* `git diff` shows no change under `cmd/`, `internal/`, `.github/`, `go.mod`, `go.sum`, `Makefile`, or `selfupdate-release.json`.
* No commit and no tag in this plan.

## Execution record 2026-11-18

The owner accepted the MADR and approved this plan on 2026-11-18. The root `README.md` was replaced. The seven guides were added. `docs/architecture.md` was rewritten. `docs/README.md` keeps the hand-written index, including the `0009` rows, and its "I want to…" table starts with the same eight tasks as the root README.

Prose lines were wrapped to satisfy markdownlint MD013 (200 columns). Code fences and tables were left long. No fact from the phases was dropped to make a line shorter.

Planted checks, outside this tree, before the real tree was called clean:

* A copy of the finished root `README.md` with `read_url` inserted into the overview failed the content check with `banned read_url`. The real file passed: both installer one-liners and `search_books` were present, and the banned strings were absent.
* A filesystem copy of the tree with `[missing](does-not-exist.md)` appended to `docs/guides/installing.md` made `python3 scripts/check_records.py --check-all` exit 1 and print `broken relative link: docs/guides/installing.md:77: does-not-exist.md`. The copy was deleted.

On the real tree, after the last edit:

* `python3 scripts/check_records.py --check-all` exited 0.
* `npx --yes markdownlint-cli2@0.23.2` exited 0. It linted 40 files and reported 0 issues.
* `python3 ../go-shared/scripts/check_disclosure.py --repo .` exited 0. The sibling checkout was `3e92bdae9c3e8fd6c662f891e803a1135ca8b513`. The summary was `disclosure: 104 files scanned, 0 finding(s), 0 unreadable`. That command lists tracked files only, so the new guides were not in those 104. A scratch repository containing the 12 new and rewritten Markdown files was scanned with the same script and printed `disclosure: 12 files scanned, 0 finding(s), 0 unreadable`, then the scratch repository was deleted.

`git diff --name-only` and `git status` show `README.md`, `docs/README.md`, `docs/architecture.md`, the `0009` pair, and `docs/guides/`. No change under `cmd/`, `internal/`, `.github/`, `go.mod`, `go.sum`, `Makefile`, or `selfupdate-release.json`. No commit and no tag.

### Deviation 2026-11-18 — source of truth

The owner stated that this repository is the main source of truth. Phase 1 had kept the mirror notice. That notice is removed from `README.md`. The root page and `docs/architecture.md` state that this repository is the source of truth. The scaffold bullet at `HEAD` in `docs/architecture.md` that called the tree a published export of a private source is not carried forward. The historical sentence in [0006-PLAN-scaffold-agent-workspace.md](0006-PLAN-scaffold-agent-workspace.md) stays, with that clause marked withdrawn.

### Deviation 2026-11-18 — quick start

The owner asked for the one-line installs on the root README and a brief quick start section. `## Install` already held the two latest-release commands. `## Quick start` now comes first and repeats those commands, then the no-argument launch and the MagicTools and standalone client sentences. The table of contents lists the new heading.
