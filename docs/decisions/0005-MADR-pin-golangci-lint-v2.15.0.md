---
status: accepted
date: 2026-11-17
---
# Pin golangci-lint to v2.15.0 so CI can read Go 1.27.4 export data

## Context and Problem Statement

[0002-MADR-pin-go-1.27.4.md](0002-MADR-pin-go-1.27.4.md) sets this module's `go` line to 1.27.4. CI installs that toolchain from `go.mod` and then runs:

```text
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.1
GOLANGCI_LINT="$(go env GOPATH)/bin/golangci-lint" make lint
```

That line is in `.github/workflows/ci.yml`. `make lint` runs `golangci-lint run -c .golangci.yml ./...`.

Push `1e8c0ee` failed CI run [44120982374](https://github.com/maccavelli/mcp-server-duckduckgo/actions/runs/44120982374). `go test ./...` and `go vet ./...` passed. The lint step failed with:

```text
internal/models/models.go: could not import fmt (... could not load export data: internal error in importing "internal/goarch" (cannot decode "internal/goarch", export data version 5 is greater than maximum supported version 4))
```

The same step failed on the previous push, run [44088246731](https://github.com/maccavelli/mcp-server-duckduckgo/actions/runs/44088246731). `build` did not start, because it needs `validate`.

Go 1.27.4 writes unified export data version V5. `internal/pkgbits` rejects a version greater than or equal to `numVersions`. The panic text is `export data version %d is greater than maximum supported version %d`.

The module proxy `go.mod` for each linter release, read on 2026-11-17:

| golangci-lint | `go` line | `golang.org/x/tools` | `internal/pkgbits` top version |
|---|---|---|---|
| v2.13.1 | 1.26.0 | v0.49.0 | V4 |
| v2.14.0 | 1.26.0 | v0.50.0 | V5 |
| v2.15.0 | 1.26.0 | v0.52.0 | V5 |

v2.15.0 is the newest release, published 2026-10-16. v2.14.0, published 2026-09-24, is the oldest of these three whose bundled `x/tools` declares V5.

A local `go install` of v2.15.0, built with go1.27.4, then `golangci-lint run -c .golangci.yml ./...`, did not report the export-data error. It reported three findings, and exited 1:

* `internal/engine/books.go:28`: gocritic `regexpSimplify` on `` `^OL[0-9]+W$` ``
* `internal/engine/test_helper_test.go:21`: unused type `errorReader`
* `internal/engine/test_helper_test.go:23`: unused method `(*errorReader).Read`

`.golangci.yml` sets `linters.default: none` and `issues.max-issues-per-linter: 0`. `exhaustive` is off, so the v2.15.0 option rename for that linter does not apply. `revive` names its rules one by one. Enabled linters that moved inside v2.14.0 or v2.15.0 include `gocritic` 0.15.0, `gosec` 2.29.0, `revive` 1.17.1, and `staticcheck` 0.8.2.

`errorReader` has no other reference. In Go's `regexp` package, `\d` is `[0-9]`. `TestBookSearchMapsWorks` already feeds the bare key `OL2W` through `olBareWorkKey`.

Which linter release does CI install?

## Decision Drivers

* CI on `main` is red at the lint step.
* `go test` and `go vet` already pass on Go 1.27.4.
* The Go floor stays 1.27.4.
* The lint config stays an explicit enable list. A green run comes from code that satisfies that list.
* v2.15.0 is the current stable release and its `x/tools` dependency decodes V5.

## Considered Options

* Install `golangci-lint` v2.15.0 and clear the three findings that release reports
* Install v2.14.0, the oldest release in this set whose `x/tools` decodes V5
* Keep v2.13.1

## Decision Outcome

Chosen option: "Install `golangci-lint` v2.15.0 and clear the three findings that release reports", because that is the current release whose typechecker can read Go 1.27.4 export data, and the three findings are what that binary reports on this tree.

## Consequences

* Good, because the typecheck error that failed runs 44120982374 and 44088246731 is absent under v2.15.0.
* Good, because `.golangci.yml` can stay as it is. `exhaustive` is off, and `revive` already lists its rules.
* Neutral, because `gocritic`, `gosec`, `revive`, and `staticcheck` are newer inside v2.15.0 than inside v2.13.1. A later edit can surface a new finding from those bumps.
* Bad, because the pattern in `books.go` and the dead `errorReader` helper have to change for `make lint` to exit 0. Those edits are part of getting CI green.

### Confirmation

`.github/workflows/ci.yml` installs `github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0`. `golangci-lint` v2.15.0 `run -c .golangci.yml ./...` exits 0. `go test -count=1 ./internal/engine -run TestBookSearch` exits 0. The `go` line remains 1.27.4.

## Pros and Cons of the Options

### Install `golangci-lint` v2.15.0 and clear the three findings that release reports

* Good, because v2.15.0 is the newest stable tag and bundles `x/tools` v0.52.0, which declares V5.
* Good, because a local run on 2026-11-17 already showed the export-data error gone and named the only three findings.
* Neutral, because `revive` 1.17.1 describes several enabled rules as firing more consistently. This tree did not trip them in that run.

### Install v2.14.0, the oldest release in this set whose `x/tools` decodes V5

* Good, because v2.14.0 is the first tag in the table whose `pkgbits` declares V5.
* Bad, because v2.15.0 is already published, and this repo would pin one release behind the current linter on the day of the change.
* Neutral, because `gocritic` 0.15.0 arrives in v2.14.0, so the `regexpSimplify` finding still has to be cleared.

### Keep v2.13.1

* Good, because the workflow file stays unchanged.
* Bad, because v2.13.1 bundles `x/tools` v0.49.0, whose `pkgbits` stops at V4, and CI stays red on Go 1.27.4.

## More Information

The `Makefile` comment that says "MADR 0005" describes the build-kind rule from mcplib. This record is this repository's decision 0005. The comment stays.

Implementation plan: [0005-PLAN-pin-golangci-lint-v2.15.0.md](0005-PLAN-pin-golangci-lint-v2.15.0.md).

The Go floor stays the decision in [0002-MADR-pin-go-1.27.4.md](0002-MADR-pin-go-1.27.4.md).
