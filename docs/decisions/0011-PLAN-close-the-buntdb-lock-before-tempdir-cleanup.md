---
status: "implemented"
date: 2026-11-20
associated-madr: "0011-MADR-close-the-buntdb-lock-before-tempdir-cleanup.md"
---
<!-- markdownlint-disable MD013 MD024 -->

# Close the BuntDB lock file before TempDir cleanup

Associated MADR: [0011-MADR-close-the-buntdb-lock-before-tempdir-cleanup.md](0011-MADR-close-the-buntdb-lock-before-tempdir-cleanup.md)

## Goal

`Store.Close` releases (and mcp-buntdb removes) the advisory `.db.lock`
handle, so `t.TempDir()`'s RemoveAll cleanup cannot fail on windows; the
twelve buntdb-backed tests in `internal/engine` pass on windows/amd64
(mac420-laptop), and the linux suites stay green.

## Grounding Reference

* [0011-MADR-close-the-buntdb-lock-before-tempdir-cleanup.md](0011-MADR-close-the-buntdb-lock-before-tempdir-cleanup.md)
  (this decision: option A — `Store` carries `dbPath`; `Close` releases the
  lock through mcp-buntdb's `ReleaseDBLock`).
* mcp-buntdb `store.go` — `OpenBuntDB` keeps the flock handle in its
  package-level `lockFiles` map; `ReleaseDBLock(dbPath)` is its documented
  shutdown call ("closes and removes the advisory lock file"); it refuses
  `":memory:"` itself.
* The red evidence: the twelve windows failures captured on mac420-laptop at
  `05474ff` (every failure lands in `testing.go` TempDir cleanup on the still
  -held `scraping.db.lock`), never seen on linux CI.
* The owner's ruling of 2026-11-20: "draft the pair, approved to fix".

## Prerequisites

* go1.27.4 on mac420-laptop and here; the scratch clone for windows-first
  verification exists on mac420-laptop (`~/gitrepos/mcp-server-duckduckgo`).
* The red state: the twelve windows failures at `05474ff` are the recorded
  reproduction; linux CI (the 05474ff push, run green on push) never saw
  them.

## Phase 1 — the fix, test-first

1. The red state is the existing twelve windows failures — recorded here as
   the reproduction; no new test needs to go red first.
2. `internal/store/store.go`: `Store` gains `dbPath string`; `Open` sets it
   (and the in-memory fallback carries `":memory:"`, which mcp-buntdb's
   `ReleaseDBLock` already refuses); `Store.Close` calls
   `mcpbuntdb.ReleaseDBLock(s.dbPath)` after closing the database and still
   returns the close's error.
3. A regression test in `internal/store/store_test.go`. ~~After `s.Close()`,
   the `.db.lock` file is gone from the temp dir (mcp-buntdb removes the
   advisory lock file on release).~~ Deviation 2026-11-20: mcp-buntdb v1.0.1
   `ReleaseDBLock` closes the handle and does not delete the file
   (`store.go:81-91`). The test writes `k`=`v`, closes, asserts
   `scraping.db.lock` is still present, opens the same directory, and reads
   `v`. A held lock makes the second `OpenBuntDB` degrade to memory and the
   key is missing. It is red on a scratch copy of HEAD and green with this
   `Close`.

## Phase 2 — fences and gates

1. mac420-laptop windows-first: `go test ./...` green after the fix; the
   twelve failure lines gone.
2. The session host: the same suites green — no linux regression.
3. `make check-records markdownlint` pass; `docs/README.md` carries the 0011
   rows.

## Rollback

Revert the fix commit. The tests return to the red windows state; no data,
config or CI state changes.

## Completion Criteria

1. The twelve windows failures' line no longer appears in windows runs
   (mac420-laptop, `go test ./...` green).
2. A linux run of the same suites stays green (no regression).
3. The records + `docs/README.md` carry the pair and the evidence.

## Execution record

### 2026-11-20 — Phase 1 and Phase 2

`Store` carries `dbPath`. `Open` sets it to the `scraping.db` path. The in-memory fallback sets `":memory:"`. `Close` calls `mcpbuntdb.ReleaseDBLock` after `s.db.Close()` and returns that close error.

Deviation, same date. mcp-buntdb v1.0.1 `ReleaseDBLock` (`store.go:81-91`) closes the handle and does not delete `scraping.db.lock`. The plan's "file is gone" assertion failed on the fixed tree as well as on HEAD:

```text
--- FAIL: TestCloseReleasesLockFile (0.01s)
    store_test.go:66: lock file still present after Close; TempDir cleanup on windows will fail: <nil>
```

The replacement test writes `k`=`v`, closes, asserts the lock file remains, and reads `v` from a second `Open`. On a scratch clone of `05474ff` with only that test added, the second open logged `resource temporarily unavailable` and failed:

```text
--- FAIL: TestCloseReleasesLockFile (0.01s)
    store_test.go:85: second open did not read the file db (held lock degrades to memory): got "" err not found
```

The same test on the fixed `store.go` exited 0. `go test ./internal/store -count=1` on the session host exited 0.

`go test -count=1 -timeout 30m ./...` on the session host through `run-heavy` exited 0. Eleven packages reported `ok`, including `internal/engine` (12.925s) and `internal/store` (0.229s).

The linux run was `Ubuntu-24.04` through `run-on-host --need linux`. It exited 0. Eleven packages reported `ok`, including `internal/engine` (12.339s) and `internal/store` (0.029s).

`run-on-host --need windows` did not run the suite on windows. Remote tar exited 2 on `./AGENTS.md: Cannot create symlink to CLAUDE.md`. That dispatch then ran the suite on the session host. That result is not the windows result. The suite was packed again without `AGENTS.md` and run on mac420-laptop under that host's `run-heavy`. `go version` there printed `go version go1.27.4 windows/amd64`. The suite exited 0. Eleven packages reported `ok`, including `internal/engine` (12.230s) and `internal/store` (0.054s). No `scraping.db.lock` TempDir line appeared.

`python3 scripts/check_records.py --check-all` exited 0. `npx --yes markdownlint-cli2@0.23.2` exited 0 (0 issues). `docs/README.md` carries the 0011 rows.
