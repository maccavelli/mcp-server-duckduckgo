---
status: accepted
date: 2026-11-20
---
# Close the BuntDB lock file before TempDir cleanup so windows tests pass

## Context and Problem Statement

The first windows/amd64 test run of this repository (mac420-laptop,
go1.27.4, at `05474ff`) failed twelve buntdb-backed tests/subtests
(`TestBookSearchMapsWorks`, `TestBookSearchLimitClamp` ×3,
`TestBookSearchBlankQuerySkipsNetwork`, `TestBookSearchContactUserAgent`,
`TestBookSearchErrors` ×5, `TestNewSearchEngine`, `TestNewRequest`) with:

```text
testing.go:1619: TempDir RemoveAll cleanup: unlinkat
C:\...\Temp\...\001\scraping.db.lock: The process cannot access the file
because it is being used by another process.
```

No test assertion failed; the failure is `t.TempDir()`'s RemoveAll cleanup.
The lock file's open handle belongs to mcp-buntdb's `OpenBuntDB`, which
flocks `dbPath + ".lock"` and keeps the handle in a package-level map
(`lockFiles`, mcp-buntdb `store.go:16,51`); closing the buntdb database
(`Store.Close` → `s.db.Close()`) closes the database file but never the
wrapper's lock-file handle. `mcp-buntdb` exposes the release path —
`ReleaseDBLock(dbPath)`, "call this when shutting down a store opened via
OpenBuntDB" — and `Store.Close` (internal/store/store.go) never calls it. On
linux, RemoveAll of an open file succeeds, so CI on ubuntu never saw this;
the first windows/amd64 run (mac420-laptop, the host's fresh `05474ff`
clone) exposed it.

## Decision Drivers

* Windows refuses to delete an open file; the test tree must release every
  buntdb lock handle before `t.TempDir()` cleanup runs.
* `Store.Close` already cancels the shrink goroutine (`s.cancel()`) and
  closes the database; the missing piece is mcp-buntdb's documented shutdown
  call `ReleaseDBLock(dbPath)`, which closes and removes the advisory lock
  file.
* The failing tests are the red state: every buntdb-backed test in
  `internal/engine` fails on windows at `05474ff`; on linux the same suite
  is green, so the fix must be verified windows-first (mac420-laptop).
* The records flow: one short record pair per repository, approved through
  CI, per 0061-MADR's precedent and the owner's ruling of 2026-11-20
  ("approved to fix").

## Considered Options

* **A. `Store.Close` calls `ReleaseDBLock(s.dbPath)`; `Store` carries the
  db path.** Good, because the lock file's handle is released (and removed)
  wherever the store shuts down — production shutdown included, not just
  tests; the in-memory fallback's `:memory:` path is already guarded in
  `ReleaseDBLock`. Bad, because `Store` gains a `dbPath` field.
* **B. Each test registers its own `t.Cleanup` closing the store before
  TempDir teardown.** Good, because it is test-side only. Bad, because it
  must be repeated in every test file (client_test.go:21 does it for one
  store; the failing tests span two packages), the ordering depends on
  registration order, and production shutdown still leaks the lock handle.
* **C. Move the lock file outside the temp dir.** Good, because cleanup
  never sees it. Bad, because it changes the store's layout and the lock's
  path contract for a cleanup bug.

## Decision Outcome

Chosen option: "A" — `Store` carries its `dbPath`; `Store.Close` calls
mcp-buntdb's `ReleaseDBLock(s.dbPath)` after closing the database (the
in-memory fallback's `:memory:` path is already guarded in `ReleaseDBLock`).
Steps and gates in
[0011-PLAN-close-the-buntdb-lock-before-tempdir-cleanup.md](0011-PLAN-close-the-buntdb-lock-before-tempdir-cleanup.md).

* Good, because the lock file's handle (and, per mcp-buntdb's contract, the
  lock file itself) is released wherever a store shuts down — production
  shutdown included, not just tests.
* Good, because the twelve windows failures are the already-red state; the
  fix is verified windows-first on mac420-laptop, linux re-run as the
  regression check.
* Bad, because `Store` gains a `dbPath` field and `Open`'s in-memory
  fallback must carry its `:memory:` marker the same way.

## Consequences

* Windows test runs (the 0063 "pseudo-console" flow's host) stop failing in
  TempDir cleanup; the suite's first windows run after the fix is the
  recorded green.
* The store's shutdown releases the advisory lock handle even outside tests.
  ~~A shut-down server leaves no stale `.db.lock`.~~ The file remains,
  unlocked. See the amendment.
* `internal/store`'s `Store` struct gains `dbPath`; `Open`'s in-memory
  fallback carries `":memory:"`, which `ReleaseDBLock` already refuses.

## Confirmation

* On mac420-laptop: `go test ./...` green after the fix, windows-first.
* On the session host and CI (ubuntu): the same suite green — no linux
  regression.
* The twelve named tests' failure line no longer appears in windows runs.

## Amendment 2026-11-20

mcp-buntdb v1.0.1 `ReleaseDBLock` (`store.go:78-91`) comments that it
closes and removes the advisory lock file. The body closes the handle and
deletes the map entry. It does not delete `scraping.db.lock`. Option A is
unchanged: `Store.Close` calls `ReleaseDBLock` after `s.db.Close()`. The
Windows failure is the open handle, and TempDir `RemoveAll` can delete the
file once that handle is closed. A shut-down server leaves the lock file in
place, unlocked. `TestCloseReleasesLockFile` writes a key, closes, asserts
the lock file remains, and reads the key back from a second `Open`. On a
scratch copy of `05474ff` that second open logged `resource temporarily
unavailable` and the key was missing. The same test passed after this
`Close`.
