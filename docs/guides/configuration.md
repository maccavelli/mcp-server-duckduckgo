# Configuration

Task index: [docs/README.md](../README.md).

`serve`, and any command that is not marked `selfupdate.skip-config`, calls `config.Load`. `update` and `identity` carry that annotation, so they do not create the cache directory or `config.yaml`.

## Cache directory

`config.Load` and `serve` use `os.UserCacheDir()` joined with `mcp-server-duckduckgo`.

| OS | Directory |
| :--- | :--- |
| Linux | `$XDG_CACHE_HOME/mcp-server-duckduckgo`, or `$HOME/.cache/mcp-server-duckduckgo` when `XDG_CACHE_HOME` is unset |
| macOS | `$HOME/Library/Caches/mcp-server-duckduckgo` |
| Windows | `%LocalAppData%\mcp-server-duckduckgo` |

Files in that directory:

| File | When it appears |
| :--- | :--- |
| `config.yaml` | `config.Load` creates it with a comment template when the file is absent. |
| `run.lock` | `serve` creates it unless `MCP_ORCHESTRATOR_OWNED=true`. |
| `scraping.db` | `serve` opens this BuntDB file for VQD tokens, cookies, user-agent rotation, and provider health. If the file cannot be opened, the store falls back to memory for that process. |

`config.yaml` has no keys the server reads. `config.GetString` is called only from `internal/config/config_test.go`.

## Environment

| Variable | Effect |
| :--- | :--- |
| `DDG_DEFAULT_FORMAT` | Format used by the search and media tools when the call omits `format`. Empty means `json`. |
| `OPENLIBRARY_CONTACT` | When non-empty after trimming, appended to the Open Library user agent. Empty keeps `mcp-server-duckduckgo (+https://github.com/maccavelli/mcp-server-duckduckgo)`. |
| `GH_TOKEN`, `GITHUB_TOKEN` | Sent only to the GitHub API origin by `update`, to raise the rate limit. |
| `MCP_ORCHESTRATOR_OWNED` | `true` makes `serve` skip the singleton lock. |
| `GOMEMLIMIT` | `serve` sets `1024MiB` only when the variable is unset. |
| `GOMAXPROCS` | `serve` sets `2` only when the variable is unset. |

## Page fetch

`internal/engine/fetch.go` retries with the HTTP client. When the
response status is 403 or the body is classified as a challenge, fetch
tries Playwright. `playwright.Install` runs on that path, once per process.
A failed install or launch is logged, and the HTTP result stands.
Do not install a browser as a setup step before the first search.
