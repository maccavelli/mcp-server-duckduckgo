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

## I want to…

| I want to… | Start here |
| :--- | :--- |
| understand how this repository is put together | [architecture.md](architecture.md) |
| know why `search_books` left Anna's Archive | [0001-MADR](decisions/0001-MADR-retarget-book-search-open-library.md) |
| implement or verify the Open Library retarget | [0001-PLAN](decisions/0001-PLAN-retarget-book-search-open-library.md) |
| know why the module declares `go 1.26.11` with `toolchain go1.27.4` | [0007-MADR](decisions/0007-MADR-move-toolchain-to-go-1-27-4-and-floor-to-1-26-11.md) |
| read the earlier decision that set the floor to Go 1.27.4 (history) | [0002-MADR](decisions/0002-MADR-pin-go-1.27.4.md) |
| pin Go 1.27.4 and rebuild `dist/` | [0002-PLAN](decisions/0002-PLAN-pin-go-1.27.4.md) |
| know which dependency pins moved | [0003-MADR](decisions/0003-MADR-update-dependency-pins.md) |
| update the module pins | [0003-PLAN](decisions/0003-PLAN-update-dependency-pins.md) |
| know why playwright-go uses the mxschmitt module | [0004-MADR](decisions/0004-MADR-move-playwright-to-maintained-module.md) |
| move playwright-go to the maintained module | [0004-PLAN](decisions/0004-PLAN-move-playwright-to-maintained-module.md) |
| know why CI lint fails on Go 1.27.4 | [0005-MADR](decisions/0005-MADR-pin-golangci-lint-v2.15.0.md) |
| pin golangci-lint so CI can typecheck | [0005-PLAN](decisions/0005-PLAN-pin-golangci-lint-v2.15.0.md) |
| know why the agent workspace is in this repo | [0006-MADR](decisions/0006-MADR-scaffold-agent-workspace.md) |
| install the agent workspace | [0006-PLAN](decisions/0006-PLAN-scaffold-agent-workspace.md) |
| move the floor and toolchain pins | [0007-PLAN](decisions/0007-PLAN-move-toolchain-to-go-1-27-4-and-floor-to-1-26-11.md) |
