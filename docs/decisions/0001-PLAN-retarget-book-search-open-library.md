---
status: complete
date: 2026-11-16
associated-madr: "0001-MADR-retarget-book-search-open-library.md"
---
# Implement Open Library book search

Associated MADR: [0001-MADR-retarget-book-search-open-library.md](0001-MADR-retarget-book-search-open-library.md)

## Goal

Replace Anna's Archive HTML scraping in `BookSearch` with a single `GET` to Open Library `search.json`, mapping the JSON `docs` list onto `models.SearchResult`, so `search_books` returns book metadata without a DDoS-Guard challenge.

## Scope

In scope:

* `SearchEngine.BookSearch` and its tests
* MCP tool description for `search_books`
* Removing Anna's Archive host construction and HTML helpers used only by book search

Out of scope:

* Membership keys, Playwright challenge solving, or any Anna's Archive request
* Google Books, OpenAlex, Crossref, or a multi-provider book runner
* Changing the `BookSearch` method signature or the `search_books` tool name
* Fetching Open Library cover bytes or edition sub-objects (`editions` field)
* Full-module `go test ./...` unless a later phase names `run-on-host` / `run-heavy`

## Implementation Steps

### Phase 1 — Engine: Open Library JSON client

1. Add `internal/engine/books.go` that implements `BookSearch` as one GET. Build it with `http.NewRequestWithContext` on `e.Client`. Do not call `newRequest`: that persists a browser User-Agent for the hostname via `Store.GetUA`.

   * Trim the query. If it is empty, return `fmt.Errorf("book search query is empty")` and do not call the client.
   * URL: `https://openlibrary.org/search.json`
   * `q` = trimmed query. `limit` = `maxResults` clamped to `[1, 50]`; if `maxResults <= 0`, use 5. The handler also turns non-positive `max_results` into 5 before this method. The server default is 100 and `limit=1000` returned 1,000 docs, so the upper clamp is a client bound. `limit=0` is a legal server value that returns `docs: []` while `numFound` stays non-zero; never send it.
   * `fields` = `key,title,subtitle,author_name,first_publish_year,number_of_pages_median,cover_i,edition_count` as one comma-separated value (`parse_comma_separated_list` in `openlibrary/fastapi/models.py`). Do not send `isbn`, `publisher`, `language`, `editions`, `availability`, or `*`.
   * `Accept: application/json`
   * `User-Agent`: `mcp-server-duckduckgo (+https://github.com/maccavelli/mcp-server-duckduckgo)`. If `OPENLIBRARY_CONTACT` is non-empty after trim, append it inside the parenthesis, separated by `; `. Do not read `main.RawVersion` and do not invent an email.
   * Decode with `json.NewDecoder(io.LimitReader(resp.Body, config.MaxBodyBytes))` into a struct that has `docs` only. Ignore envelope counts. `number_of_pages_median` is `float64` (a fractional JSON number into `int` fails the whole document on go 1.27.4). `cover_i`, `edition_count`, and `first_publish_year` are `int`.
   * Non-200, `isJSONEndpointServingHTML`, or decode error: return a wrapped error that names Open Library and the status or cause. Do not retry on 429. An unknown `fields` name is HTTP 200 with empty doc objects, not 422; those docs fail the title/key check below.
   * No doc survives mapping: `fmt.Errorf("book search returned no results")`.

2. Map each doc. Skip it when `title` is empty or the key is not a work key.

   * Work URL: `https://openlibrary.org` + key when the key starts with `/works/`. A bare key matching `OL` + digits + `W` becomes `https://openlibrary.org/works/` + key. Any other key is skipped. `https://openlibrary.org/OL30906747W` is a proven 404; `https://openlibrary.org/works/OL30906747W` is a proven 200.
   * `Title`: `title`, then `: ` + `subtitle` when subtitle is non-empty.
   * `Author`: `author_name` joined with `", "`.
   * `Publisher`: empty. The first of 113 publishers on `/works/OL27448W` was `Hör Verlag, München`.
   * `Date`: `first_publish_year` as a decimal string when non-zero.
   * `Thumbnail`: `https://covers.openlibrary.org/b/id/{cover_i}-M.jpg` when `cover_i` > 0. Construct the URL. Do not GET it.
   * `Source`: `Open Library`.
   * `Info` and `Description`: the same middot-joined line, parts omitted when absent — authors, year, median pages (integer when the float is whole), `{n} editions`. `SearchResponse.ToMarkdown` quotes `Description` and does not print `Author` or `Info`, so the line has to live in `Description`. Do not append a language code.
   * Truncate every mapped string with `truncate(..., config.MaxSnippetLength)`.

3. Delete `internal/engine/mirror.go`. Do not leave unused `queryMirror` / `parseMirrorHTML` / `extractMirrorResult` / `fetchMirrorHTML`.

4. Replace `internal/engine/mirror_test.go` with `internal/engine/books_test.go`:

   * Fixture: a 200 JSON body with `"offset": null`, both count fields, and two `docs`. One key is `/works/OL1W` with subtitle, authors, year, whole-number `number_of_pages_median`, `cover_i`, and `edition_count`. One key is bare `OL2W` with title only. One extra doc has an empty title and must be dropped.
   * Assert one GET, host `openlibrary.org`, path `/search.json`, `q` and `limit` set, `fields` equal to the list above (no `isbn`, `publisher`, or `language`), `Accept: application/json`, and `User-Agent` equal to the constant (not a `Mozilla` string).
   * Assert the `/works/` URL, the bare key rewritten to `/works/OL2W`, source, thumbnail, date, empty publisher, and that `Description` equals `Info` and contains the year.
   * `maxResults` 0 and -1 send `limit=5`. `maxResults` 100 sends `limit=50`.
   * Blank and whitespace queries return an error and the test server receives no request.
   * Cases: non-200; `content-type: text/html`; truncated JSON; empty `docs`; a 200 whose only doc has no title.
   * **Fail-first:** on a scratch copy (temp file or throwaway clone, not this tree), run the truncated-JSON case against the old AA scraper or against an empty `BookSearch` and record the failure string in the phase handoff. Then land the assertion here.

5. `gofmt` the new files. Session-host `go test ./internal/engine` (single package; no `run-on-host`).

### Phase 2 — Tool copy and leftover AA references

1. Update the `search_books` description in `internal/handler/search/tools.go` so it describes an Open Library work-catalog lookup (title, author, year). Keep the tool name `search_books`. Drop wording that implies ISBN resolution, shadow-library retrieval, or Anna's Archive.
2. Update the `SearchResult` comment in `internal/models/models.go` if it still says "or mirrors".
3. Grep the Go tree for `annas-archive`, `js-aarecord`, `queryMirror`, and `BookSearch performs a book search using Anna`. Those strings must be gone from `*.go`.
4. Session-host `go test ./internal/engine ./internal/handler/search`.

### Phase 3 — Verification and handoff

1. Re-read `internal/engine/books.go` in full (not a diff glance) and confirm the URL, field list, User-Agent, and mapping match this plan and the MADR.
2. If a live check is wanted in the same turn: one GET to the probe URL from Phase 1's field list, confirm HTTP 200 and non-empty `docs`. Do not add that call to CI.
3. Record execution in this PLAN (what landed, test output, any deviation). Set this PLAN `status: complete` only when every acceptance criterion below is met. Set the MADR `status: accepted` when the owner has approved the decision (that may already have happened via `proceed`).

## Verification

Acceptance:

* `go test ./internal/engine` passes on the session host.
* `go test ./internal/handler/search` passes on the session host.
* `BookSearch` issues one GET to `openlibrary.org/search.json` and never to `annas-archive.*`.
* A complete fixture doc produces non-empty `Title`, `URL` with path prefix `/works/`, `Author`, `Description` equal to `Info`, empty `Publisher`, and `Source == "Open Library"`.
* The request `fields` value is `key,title,subtitle,author_name,first_publish_year,number_of_pages_median,cover_i,edition_count`.
* `limit` is 5 when `maxResults <= 0` and 50 when `maxResults > 50`. A blank query performs no HTTP request.
* Truncated JSON, non-200, and empty `docs` each return an error (the truncated-JSON gate was seen to fail on broken input on a copy before it was committed).
* `gofmt` and the repo pre-add wrapper (`make pre-add-check` / `scripts/go-precheck.sh` if present; otherwise `gofmt -l` plus `golint` on staged `.go` files) are clean for the staged Go files.

## Rollout and Rollback

Rollout is the next published binary of this MCP server. There is no schema or config migration. Operators need no new environment variable; `OPENLIBRARY_CONTACT` is optional.

Rollback is reverting the engine/handler commit. That restores the Anna's Archive scraper, which is known-403 from this environment; do not roll back unless a worse Open Library regression appears.

## Review 2026-11-16

Before execution. The contract probe and `go1.27.4` unmarshal check are in the MADR section "Contract probe, 2026-11-16". This plan changed as follows. No engine code has been written.

* Work URLs use `/works/`. Prefixing a bare id with only `/` produces a 404.
* `isbn`, `publisher`, and `language` are not requested. On `/works/OL27448W` those lists were 473, 113, and 18 long, and the first publisher and language were not a canonical edition.
* `limit` stays clamped at 50 because the server accepted `limit=1000`. `limit=0` is legal and returns an empty `docs` array with a non-zero `numFound`.
* User-Agent has no build version. `main.RawVersion` is a main-package ldflag.
* The request is not built with `newRequest`.
* `Description` carries the metadata line so markdown output shows it.
* A blank query does not hit the network.
* `number_of_pages_median` is a `float64` because a fractional number into `int` fails the document. Live medians in the probe were whole numbers.

## Execution record

2026-11-16. Owner said "approved to proceed". Phases 1–3 landed in the working tree. HEAD is `main`. This turn did not commit: branch gates require an explicit commit on `main`.

### Fail-first (scratch clone `/tmp/ol-failfirst`, not this tree)

`TestFailFirstTruncatedJSON` fed `{"docs":[` to the Anna's Archive `BookSearch` and required `unexpected end of JSON input`.

```
--- FAIL: TestFailFirstTruncatedJSON (0.01s)
    books_failfirst_test.go:26: gate not met, got: book search returned no results from any mirror
FAIL
```

The old scraper parsed that body as HTML, found no cards, and reported an empty mirror search.

### What landed

* `internal/engine/books.go` — one HTTPS GET to `https://openlibrary.org/search.json`.
* `internal/engine/books_test.go` replaces `mirror_test.go`.
* Deleted `internal/engine/mirror.go` and `internal/engine/mirror_test.go`.
* `search_books` description in `internal/handler/search/tools.go` now describes an Open Library work lookup.
* `SearchResult` comment in `internal/models/models.go` no longer says "or mirrors".
* `*.go` has no `annas-archive`, `js-aarecord`, `queryMirror`, or `BookSearch performs a book search using Anna`.

`json.Decoder` reports truncated `{"docs":[` as `unexpected EOF`. The test asserts that string. `json.Unmarshal` of the same bytes says `unexpected end of JSON input`; this client uses `Decoder`, so the assertion matches `Decoder`.

### Verification

`gofmt -l` on the four edited Go files printed nothing. `golint` on `books.go` and `books_test.go` printed nothing. `golint` on `tools.go` and `models.go` reports only comments and names that were already there.

```
ok  	github.com/maccavelli/mcp-server-duckduckgo/internal/engine	0.777s
ok  	github.com/maccavelli/mcp-server-duckduckgo/internal/engine	12.546s
ok  	github.com/maccavelli/mcp-server-duckduckgo/internal/handler/search	0.179s
```

The 0.777s run is `go test ./internal/engine -run TestBookSearch`. The 12.546s run is the whole engine package on this host (other tests retry DuckDuckGo while offline). No `run-on-host`.

Live check, not in CI: `GET https://openlibrary.org/search.json?q=python+programming&limit=1` with the plan field list returned HTTP 200, `application/json`, 369 bytes, one doc, key `/works/OL30906747W`, title `Core Python Programming`.

`internal/engine/books.go` was re-read in full. URL, field list, User-Agent, limit clamp, and mapping match this plan.
