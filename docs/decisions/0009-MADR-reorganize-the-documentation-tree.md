---
status: accepted
date: 2026-11-18
decision-makers: repository owner
consulted: "documentation-writing and madr-and-plan-writing, read 2026-11-18; this repository's README.md, docs/README.md, docs/architecture.md, scripts/check_records.py, Makefile, selfupdate-release.json, and the command, config, handler, store, and engine sources named below, read 2026-11-18; GitHub release v1.2.0 asset list from run 44307823508; go-selfupdate-lib v1.26.3 docs/guides/building-releases.md section 12 and mcplib v1.6.0 diagnostic.go and orchestrator.go, read from the module cache the same day; README and docs/README of mcp-server-recall, go-selfupdate-lib, and prepare-commit-msg, read the same day"
---
# Reorganize the documentation into a short README and a guide library

## Context and Problem Statement

The root `README.md` is the page a new reader opens, and it does not describe the server that `v1.2.0` publishes. [docs/README.md](../README.md) indexes decisions and asks only "why was this decided?". There is no `docs/guides/` directory. A person who wants to install, update, call a tool, or build the module has to reconstruct that from a README written for an earlier command surface.

Measured on 2026-11-18 against the tree at `5cfff6129cabab78f36f093a22e3bb5c1ce504fa` and release `v1.2.0`.

* **E1, the root README.** It tells the reader to move a downloaded binary to `~/.local/bin` on Linux, `/usr/local/bin` on macOS, or `%LOCALAPPDATA%\Programs\duckduckgo` on Windows. It says the server is stateless and needs no configuration file. Its environment table names `MCP_ENDPOINT_API_PORT` (default `47688`) and `MCP_LOG_LEVEL`. Its tool table names `search_suggest` and `read_url`, and it does not name `search_books` or `fetch_url`. The update example installs `v1.1.0`. The release paragraph says a tag publishes the three binaries and one `SHA256SUMS`, and it cites mcplib MADR 0005 for that rule.
* **E2, what the process actually does.** `cmd/mcp-server-duckduckgo/root.go` runs `serve` when the binary is started with no subcommand. `serve` writes the cache directory from `os.UserCacheDir()` plus `mcp-server-duckduckgo`, opens `scraping.db` there, and `config.Load` creates `config.yaml` in that directory when the file is absent. `config.GetString` has no caller outside `internal/config/config_test.go`. A search of `*.go` finds no `MCP_ENDPOINT_API_PORT` and no `MCP_LOG_LEVEL`. `internal/handler/search/tools.go` registers `search_web`, `search_news`, and `search_books`. `internal/handler/media/tools.go` registers `search_images` and `search_videos`. `FetchURLTool.Name` returns `fetch_url`. No registration returns `search_suggest` or `read_url`. `search_books` calls `https://openlibrary.org/search.json` (`internal/engine/books.go`). `update` and `identity` carry `selfupdate.skip-config` and do not load that configuration (`cmd/mcp-server-duckduckgo/update.go`, `identity.go`).
* **E3, the installer `v1.2.0` publishes.** `selfupdate-release.json` sets `installer.name` to `mcp-server-duckduckgo` and sets no hooks and no `env_prefix`. Release `v1.2.0` contains `install.sh`, `install.ps1`, the three binaries, and `SHA256SUMS`. go-selfupdate-lib `v1.26.3` `docs/guides/building-releases.md` section 12, read from the module cache, installs a Unix binary into `~/.local/bin` and a Windows binary into `%LOCALAPPDATA%\Programs\<name>`, checks `SHA256SUMS`, and runs the product's `identity` command. The one-liners in that section are `curl -fsSL https://github.com/<owner>/<repo>/releases/latest/download/install.sh | sh` and `irm https://github.com/<owner>/<repo>/releases/latest/download/install.ps1 | iex`. The name in this spec makes the Windows folder `mcp-server-duckduckgo` and the variable prefix `MCP_SERVER_DUCKDUCKGO`. Platforms in `update.go` are `linux/amd64`, `darwin/arm64`, and `windows/amd64`.
* **E4, how the neighbouring repositories split the same job.** `mcp-server-recall` puts the two installer one-liners, a short overview, a table of contents, and an "I want to" table in the root `README.md`, and it keeps install, client, configuration, CLI, and tool procedures in `docs/guides/`. `prepare-commit-msg` puts its installer one-liner in the root `README.md` and sends the options to the following section. `go-selfupdate-lib` keeps a task-phrased "I want to…" table in `docs/README.md` that links `docs/guides/`. `documentation-writing` fixes this repository's tree: one `docs/` for the product, unnumbered files in `docs/guides/`, and `docs/README.md` as the table of contents plus the "I want to…" matrix. This repository's `scripts/check_records.py` says it does not generate that index.
* **E5, the current docs index.** `docs/README.md` has twenty "I want to…" rows. Each row asks why a decision was made or how to carry a decision out. None names install, update, a tool call, or a build. `docs/architecture.md` says `docs/guides/` and `docs/reports/` appear with the first such document, and under "What is not here" it says the gates run by hand and CI is absent. `.github/workflows/ci.yml` runs `make verify` on a push. [0006-MADR-scaffold-agent-workspace.md](0006-MADR-scaffold-agent-workspace.md) left `docs/guides/` uncreated so the directory would not be empty.

The reader cannot install the published release, call the current tools, or build the module from the page that claims to explain those tasks.

## Decision Drivers

* The owner asked for a user-guide library, a table of contents, an "I want to…" section, the one-line installer scripts in the root `README.md`, and a brief overview.
* The guides have to match the code and `v1.2.0`, including both the person who runs the binary and the person who builds it.
* One documentation tree. Search, fetch, and the book catalog are the product.
* `documentation-writing` chooses the directories. This record chooses what the root `README.md` keeps and which guides exist.
* A statement that names a tool, a path, an environment variable, or an IDE file has to be traceable to a source read for this decision. An unverified client-file path is not copied forward.
* Records `0001` through `0008` stay as written. `0006` still forbids an empty `docs/guides/`.

## Considered Options

* Short root README, task guides, and a task matrix
* Leave the long README and add guides beside it
* Pointer README with every procedure under `docs/`

## Decision Outcome

Chosen option: "Short root README, task guides, and a task matrix", because the owner asked for the installers and the overview on the root page, and the procedures are long enough to be guides.

The root `README.md` keeps a one-paragraph overview, the two installer one-liners for the latest release, a table of contents, a short "I want to…" table, and the three `update` command lines. The IDE catalog, the manual `mv` install, the environment table, and the tool table leave that file.

Amendment 2026-11-18: the owner stated that this repository is the main source of truth. The root page does not keep the mirror notice. It states that this repository is the source of truth. `docs/architecture.md` states the same. The scaffold bullet that called this tree a published export of a private source is not carried forward.

Amendment 2026-11-18: the owner asked for the one-line installs on the root README and a brief quick start. `## Quick start` is the first procedure section. It repeats the two latest-release commands, then the launch `mcp-server-duckduckgo` with no arguments. MagicTools launches that binary in production. A standalone client uses the same command. `## Install` keeps the two commands as well.

`docs/guides/` holds seven guides, each with real steps: `installing.md`, `updating.md`, `connecting-a-client.md`, `mcp-tools.md`, `configuration.md`, `developing.md`, and `releasing.md`. The first five are for someone running the binary. The last two are for someone changing it. `docs/architecture.md` is rewritten to the current process and stays free of history. `docs/README.md` keeps its hand-written record table and replaces the decision-only matrix with tasks that start at a guide, plus the "know why" rows that still have no guide. `docs/reports/` stays absent until a report exists.

`connecting-a-client.md` documents MagicTools as the production client and the standalone stdio command. It does not catalogue IDE configuration paths from the current README. Those paths were not checked against each IDE's current docs, and several of them point at install directories the `v1.2.0` installer does not use.

### Confirmation

[0009-PLAN-reorganize-the-documentation-tree.md](0009-PLAN-reorganize-the-documentation-tree.md) is the completion check. `python3 scripts/check_records.py --check-all` must exit 0, `npx --yes markdownlint-cli2@0.23.2` must exit 0, and the disclosure check must report 0 findings. A scratch copy with one broken guide link must fail `--check-all` before the real tree is called clean.

## Consequences

* Good, because the root page shows the install command the release actually ships.
* Good, because a tool name, a cache path, and an environment variable in a guide come from the sources in E2 and E3.
* Good, because `docs/README.md` answers "I want to install" as well as "I want to know why".
* Bad, because the IDE file paths in the current README are removed rather than corrected, so a reader who used one of those examples has to take the path from that IDE's own docs.
* Bad, because `docs/architecture.md` changes in place. Its old sentences about a hand-run gate and a missing CI workflow stop being the description of the tree.
* Neutral, because records `0001` through `0008` stay historical, including sentences that describe the README and the workflow as they were on the day each record was accepted.

## Pros and Cons of the Options

### Short root README, task guides, and a task matrix

The shape used by `mcp-server-recall` for a product binary, with this repository's installer contract from E3 and this repository's one-tree rule.

* Good, because the owner asked for the one-liners and the overview on the root page, and for the procedures in a guide library.
* Good, because `check_records.py --check-all` already checks links in `README.md`, `docs/README.md`, `docs/architecture.md`, and `docs/guides/`.
* Bad, because two "I want to…" tables exist, and they drift if a later edit updates only one.

### Leave the long README and add guides beside it

* Good, because nothing already linked from outside this repository breaks.
* Bad, because the root page would still tell a reader to install into `/usr/local/bin` and to call `read_url`.
* Bad, because the same procedure would live in two places.

### Pointer README with every procedure under `docs/`

* Good, because one file would own every task.
* Bad, because the owner asked for the installer one-liners and the overview in the root `README.md`.

## More Information

`0006` created the docs tree and left `docs/guides/` for the first real guide. This record is that guide set. It does not change `go.mod`, the Makefile recipes, `.github/workflows/ci.yml`, or the release spec. It does not add `CHANGELOG.md`, `SECURITY.md`, `CONTRIBUTING.md`, or `CODEOWNERS`.

Amendment 2026-11-18. The first draft of this outcome kept the mirror notice that was already the first block of `README.md`. The owner rejected that notice on the same day: this repository is the source of truth, and changes are made here. [0006-PLAN-scaffold-agent-workspace.md](0006-PLAN-scaffold-agent-workspace.md) keeps its historical scaffold sentence and marks the export clause withdrawn.

The Windows installer variable prefix is the library's default for the name `mcp-server-duckduckgo`: upper case, with each character outside `A`–`Z` and `0`–`9` replaced by `_`, which is `MCP_SERVER_DUCKDUCKGO`. The spec does not set `env_prefix`.
