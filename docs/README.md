# Documentation

Records for this repository. The numbering sequence is repository-wide.

| Number | Kind | Title |
| :--- | :--- | :--- |
| [0001](decisions/0001-MADR-retarget-book-search-open-library.md) | MADR | Retarget book search from Anna's Archive to Open Library Search API |
| [0001](decisions/0001-PLAN-retarget-book-search-open-library.md) | PLAN | Implement Open Library book search |
| [0002](decisions/0002-MADR-pin-go-1.27.4.md) | MADR | Pin this module to Go 1.27.4 |
| [0002](decisions/0002-PLAN-pin-go-1.27.4.md) | PLAN | Implement the Go 1.27.4 pin and rebuild |
| [0003](decisions/0003-MADR-update-dependency-pins.md) | MADR | Update dependency pins to current stable releases |
| [0003](decisions/0003-PLAN-update-dependency-pins.md) | PLAN | Implement the dependency pin update |
| [0004](decisions/0004-MADR-move-playwright-to-maintained-module.md) | MADR | Move playwright-go onto the maintained module path |
| [0004](decisions/0004-PLAN-move-playwright-to-maintained-module.md) | PLAN | Implement the playwright module-path move |
| [0005](decisions/0005-MADR-pin-golangci-lint-v2.15.0.md) | MADR | Pin golangci-lint to v2.15.0 so CI can read Go 1.27.4 export data |
| [0005](decisions/0005-PLAN-pin-golangci-lint-v2.15.0.md) | PLAN | Implement the golangci-lint v2.15.0 pin |
| [0006](decisions/0006-MADR-scaffold-agent-workspace.md) | MADR | Install the fleet agent workspace |
| [0006](decisions/0006-PLAN-scaffold-agent-workspace.md) | PLAN | Implement the fleet agent workspace |
| [0007](decisions/0007-MADR-move-toolchain-to-go-1-27-4-and-floor-to-1-26-11.md) | MADR | Declare `go 1.26.11` as the floor and `toolchain go1.27.4` as the build toolchain |
| [0007](decisions/0007-PLAN-move-toolchain-to-go-1-27-4-and-floor-to-1-26-11.md) | PLAN | Implement `go 1.26.11` as the floor and `toolchain go1.27.4` |
| [0008](decisions/0008-MADR-adopt-go-selfupdate-lib-v1.26.3.md) | MADR | Adopt go-selfupdate-lib v1.26.3 for the update command and the release workflow |
| [0008](decisions/0008-PLAN-adopt-go-selfupdate-lib-v1.26.3.md) | PLAN | Implement go-selfupdate-lib v1.26.3 for the update command and the release workflow |
| [0009](decisions/0009-MADR-reorganize-the-documentation-tree.md) | MADR | Reorganize the documentation into a short README and a guide library |
| [0009](decisions/0009-PLAN-reorganize-the-documentation-tree.md) | PLAN | Implement the short README and the guide library |
| [0010](decisions/0010-MADR-adopt-the-fleet-pre-add-wrapper.md) | MADR | Adopt the fleet pre-add wrapper: golangci-lint over staged packages |
| [0010](decisions/0010-PLAN-adopt-the-fleet-pre-add-wrapper.md) | PLAN | Implement the fleet pre-add wrapper |
| [0011](decisions/0011-MADR-close-the-buntdb-lock-before-tempdir-cleanup.md) | MADR | Close the BuntDB lock file before TempDir cleanup so windows tests pass |
| [0011](decisions/0011-PLAN-close-the-buntdb-lock-before-tempdir-cleanup.md) | PLAN | Close the BuntDB lock file before TempDir cleanup |
| [0012](decisions/0012-MADR-repair-duckduckgo-news-video-and-fetch.md) | MADR | Repair DuckDuckGo news queries, video URLs, and empty page fetches |
| [0012](decisions/0012-PLAN-repair-duckduckgo-news-video-and-fetch.md) | PLAN | Implement the DuckDuckGo news, video, and fetch repair |

## I want to…

| I want to… | Start here |
| :--- | :--- |
| install the binary | [Installing](guides/installing.md) |
| update an installed binary | [Updating](guides/updating.md) |
| connect a client | [Connecting a client](guides/connecting-a-client.md) |
| call a tool | [MCP tools](guides/mcp-tools.md) |
| set the cache and environment | [Configuration](guides/configuration.md) |
| build and test | [Developing](guides/developing.md) |
| cut a release | [Releasing](guides/releasing.md) |
| understand how this repository is put together | [architecture.md](architecture.md) |
| know why `search_books` uses Open Library | [0001-MADR](decisions/0001-MADR-retarget-book-search-open-library.md) |
| know why an earlier record pinned the module to Go 1.27.4 | [0002-MADR](decisions/0002-MADR-pin-go-1.27.4.md) |
| know which dependency pins were updated | [0003-MADR](decisions/0003-MADR-update-dependency-pins.md) |
| know why playwright-go uses the mxschmitt module | [0004-MADR](decisions/0004-MADR-move-playwright-to-maintained-module.md) |
| know why CI installs golangci-lint v2.15.0 | [0005-MADR](decisions/0005-MADR-pin-golangci-lint-v2.15.0.md) |
| know why the agent workspace is in this repository | [0006-MADR](decisions/0006-MADR-scaffold-agent-workspace.md) |
| know why the module declares `go 1.26.11` and `toolchain go1.27.4` | [0007-MADR](decisions/0007-MADR-move-toolchain-to-go-1-27-4-and-floor-to-1-26-11.md) |
| know why `update` and releases use go-selfupdate-lib v1.26.3 | [0008-MADR](decisions/0008-MADR-adopt-go-selfupdate-lib-v1.26.3.md) |
| know why the root README is short and the procedures live in guides | [0009-MADR](decisions/0009-MADR-reorganize-the-documentation-tree.md) |
