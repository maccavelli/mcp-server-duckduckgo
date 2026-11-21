---
status: "implemented"
date: 2026-11-20
associated-madr: "0012-MADR-repair-duckduckgo-news-video-and-fetch.md"
---
# Implement the DuckDuckGo news, video, and fetch repair

Associated MADR: [0012-MADR-repair-duckduckgo-news-video-and-fetch.md](0012-MADR-repair-duckduckgo-news-video-and-fetch.md)

## Goal

Make `search_news` send a news.js request with `noamp=1` and a freshness `df` only when the query names one, make `search_videos` keep the watch URL `v.js` already returns, and make `fetch_url` and `fetch_content` report an empty document instead of a blank body. A news or video miss returns the `runProviders` error. `BingWebSearch` stays on `WebSearch` only.

This plan implements the recommended ruling of every open row in the MADR state table. It does not mark those rows decided. Do not edit product code until the MADR is `accepted` and the owner says to proceed.

## Grounding Reference

Snapshot at `05474ff` on 2026-11-20. `git diff` against these paths was empty: `internal/engine/news.go`, `internal/engine/video.go`, `internal/engine/fetch.go`, `internal/engine/web.go`, `internal/models/models.go`, `internal/handler/search/tools.go`, `internal/handler/media/tools.go`.

* [0012-MADR-repair-duckduckgo-news-video-and-fetch.md](0012-MADR-repair-duckduckgo-news-video-and-fetch.md), including the state table. Live calls on 2026-11-20 are the facts in Context. The `ddgs` payload maps cited there are the request and JSON shapes this plan codes against.
* `NewsSearch` calls `BingWebSearch` at `internal/engine/news.go:29-34` when `runProviders` returns an error. `ddgNewsSearch` builds the URL at `news.go:43` as `q`, `vqd`, `l=us-en`, `o=json`, `p=-1`. The excerpt is copied at `news.go:99` through `truncate` with no unescape and no tag strip. `getVQD` at `internal/engine/client.go:94` caches on the query string it is given.
* `VideoSearch` calls `BingWebSearch` at `internal/engine/video.go:28-33`. The `v.js` struct at `video.go:64-73` decodes `url`, not `content`. The result URL is `r.URL` at `video.go:86`. `TestVideoSearch/success_ddg` feeds a JSON object with `title` and `description` only.
* `internal/handler/media/tools.go:72-86` sets `page_url` from `URL` and, for videos with an empty `ImageURL`, sets `media_url` from `URL` too. `SearchMetadata` at `internal/models/models.go:24-29` has `query`, `total_count`, and `search_type`.
* `FetchContent` picks the first matching content selector at `internal/engine/fetch.go:177-185` with no text-length check. The marker `[No readable content extracted]` is assigned only at `fetch.go:207-209`. `extractStructuredText` appends ` | ` on every `td` and `th` close at `fetch.go:269-270`. `fetchResultContents` at `internal/handler/search/tools.go:145-151` returns on error and copies `Content` only when it is non-empty.
* News tool text is the `Desc` literal at `internal/handler/search/tools.go:121`. `docs/guides/mcp-tools.md` quotes that sentence in the `search_news` section.
* Helpers already in tree: `newMockClient` and `newTestEngine` in `internal/engine/test_helper_test.go`, `setupMockEngine` in `internal/engine/fetch_test.go`. `TestNewsSearch/fallback_google_news` and `TestVideoSearch/fallback_google_video` must keep passing. `TestWebSearch/fallback_bing` must keep calling Bing. `TestFetchContent_Success` must keep finding `Readable text`.

## Prerequisites

* MADR 0012 is `accepted`, and the owner has said to proceed. Until then this file stays `proposed` and no product file changes.
* Phase 1 has re-read the lines in Grounding Reference at the HEAD being edited. A moved line stops the phase. Amend this plan, then continue.
* Session host may run single-package `go test` for `./internal/engine`, `./internal/handler/search`, `./internal/handler/media`, and `./internal/models`. Do not start `go test ./...`, `make test`, or `make verify` for this work. Those are heavy runs and need `pick-host` plus `run-on-host` or `run-heavy`.
* Scratch plants go under `/tmp/ddg-0012-plant`. Delete that directory before the phase ends. Do not plant by editing the live tree and reverting.

## Phase 1 — Re-ground, then prove the new tests fail

1. Re-read the files and line numbers in Grounding Reference. Record the HEAD in the execution record. Stop if a cited line no longer matches the behavior this phase relies on.
2. On the live tree, record `git status --short` in the execution record before any copy. That list is the pre-existing dirty set. As of 2026-11-20 at `05474ff` it is `docs/README.md`, `internal/store/store.go`, `internal/store/store_test.go`, the untracked 0011 pair, and the untracked 0012 pair. Do not edit or revert those paths. Then `git clone --local . /tmp/ddg-0012-plant` so the plant is HEAD, not the dirty store files. On that copy only, add every new test named in Phases 2 through 5. Do not add the production edits there.
3. Run, from that copy, and keep the failure text. Use two groups so a missing symbol does not hide an assertion failure.

   Tests that compile against today's exported API and must fail an assertion:

   ```bash
   go test ./internal/engine -count=1 -run 'TestNewsSearch/query_shape|TestNewsSearch/excerpt_plain|TestNewsSearch/empty_query|TestNewsSearch/no_rows_skips_bing|TestNewsSearch/rows_skip_bing|TestVideoSearch/content_url|TestVideoSearch/embed_url_fallback|TestVideoSearch/no_rows_skips_bing|TestFetchContent_EmptyRoot|TestFetchContent_SkipsEmptyArticle|TestExtractStructuredText/one_cell|TestExtractStructuredText/two_cells'
   go test ./internal/handler/search -count=1 -run 'TestFetchResultContents_FetchError|TestFetchResultContents_EmptyMarker'
   go test ./internal/handler/media -count=1 -run 'TestVideoPageURL'
   ```

   Tests that name symbols this tree does not have yet (`newsToolDesc`, and field `Endpoint` on `SearchMetadata`) and must fail to compile:

   ```bash
   go test ./internal/handler/search -count=1 -run 'TestNewsEndpoint|TestNewsToolDesc'
   go test ./internal/handler/media -count=1 -run 'TestImageEndpointBingSource'
   ```

4. Every test in the first group must fail an assertion. Every test in the second group must fail to compile. A test that already passes is not a pin. Paste each failure line into the execution record. `rm -rf /tmp/ddg-0012-plant` before leaving the phase.
5. Tests that already exist (`TestNewsSearch/success_ddg`, `TestNewsSearch/fallback_google_news`, `TestVideoSearch/success_ddg`, `TestVideoSearch/fallback_google_video`, `TestWebSearch/fallback_bing`, `TestFetchContent_Success`, `TestFetchResultContents`) are not part of the plant. They stay green after the later phases.

## Phase 2 — News query, plain excerpts, no Bing web hop

Work in `internal/engine/news.go` and `internal/engine/news_test.go`.

1. Add `prepareNewsQuery(raw string) (q string, df string, err error)`.
   * Collapse whitespace with `strings.Join(strings.Fields(raw), " ")`. Do not change the case of the words that remain.
   * On that string, find these phrases case-insensitively, as whole spans bounded by the start or end of the string or by a rune that is neither a letter nor a digit. Scan runes, not bytes. Phrases, in this order: `right now` maps to `d`, `today` maps to `d`, `this week` maps to `w`, `this month` maps to `m`.
   * When a `today` span is immediately followed by `'s` or by the rune `’` plus `s`, include those two characters in the span.
   * If any span maps to `d`, `df` is `d`. Otherwise if any span maps to `w`, `df` is `w`. Otherwise if any span maps to `m`, `df` is `m`. Remove every span, then collapse whitespace again with `strings.Fields`.
   * A comma or other punctuation after a phrase stays. `rates today, extra` becomes `rates , extra`.
   * If `q` is empty, return `fmt.Errorf("news query is empty")`. Do not call `getVQD` or the network.
2. Call `prepareNewsQuery` at the start of `ddgNewsSearch`, before `getVQD`. Pass the returned `q` to `getVQD` and to news.js. The VQD cache key and the news `q` are the same string. A mismatch rejects the token.
3. Build the news.js URL with `url.Values`, not `fmt.Sprintf`. Set `q`, `vqd`, `l` = `us-en`, `o` = `json`, `p` = `-1`, `noamp` = `1`. Set `df` only when the returned `df` is non-empty. Leave `p=-1`, which is the moderate value on this endpoint.
4. Replace the excerpt assignment at `news.go:99`. Unescape with `html.UnescapeString`, strip tags with a regexp `(?i)<[^>]+>`, unescape once more, collapse whitespace with `strings.Fields`, then `truncate` to `config.MaxSnippetLength`. Put that in `plainNewsExcerpt`. Tags must not consume the 1000-rune budget.
5. Delete the `BingWebSearch` call at the bottom of `NewsSearch`. Return the `runProviders` result and error directly. Leave the Google News provider in the provider list. Do not edit `web.go`.
6. Add these subtests on `TestNewsSearch`. Use `newMockClient` and `newTestEngine`. Record every request host. The VQD response body is the text `vqd='123'`.
   * `query_shape`, table-driven. For each case the news.js query has `noamp` `1`, `p` `-1`, `l` `us-en`, `o` `json`, `q` equal to `wantQ`, and `df` equal to `wantDF`. An empty `wantDF` means the `df` key is absent. The VQD request's `q` equals `wantQ`. The `spaces` query is two spaces, `Federal`, three spaces, `Reserve`, two spaces.

   | name | query | wantQ | wantDF |
   | --- | --- | --- | --- |
   | keyword | `Federal Reserve interest rate` | `Federal Reserve interest rate` | absent |
   | today | `Federal Reserve interest rate today` | `Federal Reserve interest rate` | `d` |
   | right_now | `Federal Reserve interest rate right now` | `Federal Reserve interest rate` | `d` |
   | this_week | `Federal Reserve interest rate this week` | `Federal Reserve interest rate` | `w` |
   | this_month | `Federal Reserve interest rate this month` | `Federal Reserve interest rate` | `m` |
   | question | `what is the federal reserve doing with interest rates` | same as query | absent |
   | question_now | `what is the federal reserve doing with interest rates right now` | `what is the federal reserve doing with interest rates` | `d` |
   | leading_today | `Today in Congress` | `in Congress` | `d` |
   | possessive | `rates today's decision` | `rates decision` | `d` |
   | spaces | two spaces, Federal, three spaces, Reserve, two spaces | `Federal Reserve` | absent |
   | both_windows | `rates today this week` | `rates` | `d` |

   * `empty_query`: query `today`, query `today this week`, and query `today's`. Assert error text `news query is empty`. The mock fails the test if `Do` is called. `today's` is the possessive span from step 1, so the remaining `q` is empty.
   * `excerpt_plain`: news.js body excerpt `Tom&#x27;s <b>Fed</b>`. Assert `Description` is `Tom's Fed`.
   * `no_rows_skips_bing`: VQD succeeds, news.js returns `{"results":[]}` with status 200, Google `tbm=nws` returns status 200 and `<html><body><p>no cards</p></body></html>`. Assert a non-nil error and no request whose host contains `bing.com`.
   * `rows_skip_bing`: news.js returns one result. Assert that result is returned and no request host contains `bing.com`.
7. Leave `fallback_google_news` as it is. DDG status 500, Google `div.g` fixture, title `News 1`. That path still returns the Google row and does not call Bing.

Exit: `go test ./internal/engine -count=1 -run TestNewsSearch` passes, including the plant tests now backed by this code.

## Phase 3 — Video watch URL, no Bing web hop

Work in `internal/engine/video.go` and `internal/engine/video_test.go`. The media handler copy at `tools.go:72-86` stays. It already publishes `URL` as `page_url` and, when `ImageURL` is empty, as `media_url`.

1. Decode two more JSON fields on the `v.js` result: `content` and `embed_url`. Keep `title`, `description`, `duration`, `publisher`, and `published`.
2. Set `SearchResult.URL` from `strings.TrimSpace(content)`. When that is empty, use `strings.TrimSpace(embed_url)`. Do not prefer the `url` field when `content` is set. Keep `Duration`, `Publisher`, and `Date` as they are. Keep truncating `Description`. Do not add `description` to `MediaResult20`. That field stays off this result type.
3. Delete the `BingWebSearch` call at the bottom of `VideoSearch`. Return the `runProviders` error. Leave the Google Videos provider in the list.
4. Add subtests on `TestVideoSearch`.
   * `content_url`: `v.js` body `{"results":[{"title":"How to fold an Origami Crane","content":"https://www.youtube.com/watch?v=abc","embed_url":"https://www.youtube.com/embed/abc","duration":"3:56","publisher":"YouTube"}]}`. Assert `URL` is the watch URL, `Duration` is `3:56`, and `Publisher` is `YouTube`. Assert `URL` is not the embed URL.
   * `embed_url_fallback`: `content` absent or `""`, `embed_url` `https://www.youtube.com/embed/abc`. Assert `URL` is that embed URL.
   * `no_rows_skips_bing`: same shape as the news empty-row test, with `tbm=vid` on the Google request. Assert a non-nil error and no `bing.com` host.
5. Leave `fallback_google_video` in place. It still expects title `Video 1` from the `div.g` fixture.
6. In `internal/handler/media/tools_test.go`, add `TestVideoPageURL`. A video `MediaTool` whose `SearchFunc` returns one `SearchResult` with `URL` `https://www.youtube.com/watch?v=abc`, `Duration` `3:56`, and `Publisher` `YouTube` must respond with `page_url` and `media_url` both equal to that watch URL.

Exit: `go test ./internal/engine -count=1 -run TestVideoSearch` and `go test ./internal/handler/media -count=1 -run TestVideoPageURL` pass.

## Phase 4 — Empty documents, selectors, and layout tables

Work in `internal/engine/fetch.go`, `internal/engine/fetch_test.go`, and `internal/handler/search/tools.go` plus `tools_test.go`.

1. In the content-selector loop at `fetch.go:179-185`, skip a match whose `strings.TrimSpace(found.First().Text())` is empty. Keep the first match that has text. If every selector is empty, leave `contentNode` nil so the existing density walk runs.
2. After `extractStructuredText`, if `strings.TrimSpace(result.Content)` is empty, set `result.Content` to `[No readable content extracted]` and return a nil error. Do this before the 15,000-rune truncation. Do not call `fetchWithPlaywright` on this path. The 403 and `detectDDGChallenge` branch stays as it is.
   Deviation 2026-11-20: that assignment is the last step of `extractStructuredText`, which `FetchContent` calls before the rune truncation.
   The branch inside `FetchContent` made its cognitive complexity 51. The content and the nil error are the same.
3. In `extractStructuredText`, stop appending ` | ` on every cell close. On a `td` or `th` close, count element siblings of the parent whose tag is `td` or `th`. Append ` | ` only when that count is at least 2 and a later sibling cell exists. A one-cell row gets no separator. A two-cell row gets one separator between the cells and none after the last cell.
4. In `fetchResultContents`, when `FetchContent` returns an error, set that hit's `Content` to `err.Error()` and keep the hit. When `Content` is non-empty, including the marker, keep the existing assignment. An error wins over a body on the same return.
5. Tests:
   * `TestFetchContent_EmptyRoot`: `httptest` serves `Content-Type: text/html` and `<html><head><title>MSN</title></head><body><div id="root"></div></body>`. Assert nil error, title `MSN`, content exactly `[No readable content extracted]`.
   * `TestFetchContent_SkipsEmptyArticle`: body is `<article></article><main><p>Readable text</p></main>`. Assert the content contains `Readable text`. `TestFetchContent_Success` stays unchanged and still contains `Readable text`.
   * `TestExtractStructuredText/one_cell`: `<table><tr><td><pre>package main</pre></td></tr></table>`. Assert the result contains `package main` and does not contain ` | `.
   * `TestExtractStructuredText/two_cells`: `<table><tr><td>Left</td><td>Right</td></tr></table>`. Assert the result contains `Left | Right` and does not end with a separator.
   * `TestFetchResultContents_FetchError`: an engine whose `FetchContent` returns `errors.New("fetch failed with status 503")`. Assert the hit is still in the slice and `Content` is `fetch failed with status 503`.
   * `TestFetchResultContents_EmptyMarker`: `FetchContent` returns nil error and content `[No readable content extracted]`. Assert that string is stored on the hit.

Exit: `go test ./internal/engine -count=1 -run 'TestFetchContent|TestExtractStructuredText'` and `go test ./internal/handler/search -count=1 -run TestFetchResultContents` pass.

## Phase 5 — Endpoint name and the news tool text

1. Add field `Endpoint` to `SearchMetadata` in `internal/models/models.go`. The JSON name is `endpoint` and the tag is `omitempty`. Do not print it from `ToMarkdown`. Per-hit `source` stays the payload value.
2. In `internal/handler/search/tools.go` `Handle`, when `t.Type` is `news`, set `Metadata.Endpoint` to `DuckDuckGo`. Leave `web` and `books` empty. This labels the tool endpoint. A Google-filled news response still says `DuckDuckGo`, because this plan does not thread per-row provider identity. Image Bing fallback is out of scope and can also carry this label. The hit `source` remains `Bing` or `Bing Images` in those cases.
3. In `internal/handler/media/tools.go` `Handle`, set `Metadata.Endpoint` to `DuckDuckGo` for both images and videos.
4. Move the news `Desc` literal to `const newsToolDesc` in `internal/handler/search/tools.go`. Append this sentence before `Keywords:`: `Pass a short keyword query. The phrases today, right now, this week, and this month set freshness and are removed from the query sent to the news endpoint.` Keep the existing `[DIRECTIVE: ...]` prefix and the existing keyword list.
5. `TestNewsToolDesc` asserts `strings.Contains(newsToolDesc, "short keyword query")`. `TestNewsEndpoint` builds a news `Tool` over `mockEngine`, calls `Handle` with format `json`, and asserts `Metadata.Endpoint` is `DuckDuckGo`. The existing web `TestTool_Handle` asserts `Metadata.Endpoint` is empty.
6. `TestImageEndpointBingSource` in `internal/handler/media/tools_test.go`: image `SearchFunc` returns one hit with `Source` `Bing`, a page URL, and an image URL. Assert the media result `source` is `Bing` and `Metadata.Endpoint` is `DuckDuckGo`.
7. In `docs/guides/mcp-tools.md`, replace the quoted `search_news` description with the new `newsToolDesc` text. Under that section, add one paragraph: `search_news` sends the keyword query to DuckDuckGo news with `noamp=1`. `today` and `right now` set freshness to the past day (`df=d`). `this week` sets the past week (`df=w`). `this month` sets the past month (`df=m`). A query that is only one of those phrases returns the error `news query is empty` and does not call the network. News and video misses are errors. They are not replaced with Bing web results.

Exit: `go test ./internal/handler/search -count=1 -run 'TestNewsToolDesc|TestNewsEndpoint|TestTool_Handle'` and `go test ./internal/handler/media -count=1 -run TestImageEndpointBingSource` pass. `docs/guides/mcp-tools.md` quotes the same sentence as `newsToolDesc`.

## Phase 6 — Package tests on the session host

1. From the live tree:

   ```bash
   go test ./internal/engine -count=1
   go test ./internal/handler/search -count=1
   go test ./internal/handler/media -count=1
   go test ./internal/models -count=1
   python3 scripts/check_records.py --check-all
   npx --yes markdownlint-cli2@0.23.2
   ```

2. `gofmt -l` on the edited Go files prints nothing. Do not run `go test ./...` or `make verify` here.
3. Compare `git status --short` with the pre-existing dirty set recorded in Phase 1. Those paths may still be dirty. Do not edit them here, and do not treat them as a failure of this plan. A path that was clean in that Phase 1 status and is not in this list stops the phase: `internal/engine/news.go`, `internal/engine/news_test.go`, `internal/engine/video.go`, `internal/engine/video_test.go`, `internal/engine/fetch.go`, `internal/engine/fetch_test.go`, `internal/models/models.go`, `internal/handler/search/tools.go`, `internal/handler/search/tools_test.go`, `internal/handler/media/tools.go`, `internal/handler/media/tools_test.go`, `docs/guides/mcp-tools.md`, `docs/decisions/0012-MADR-repair-duckduckgo-news-video-and-fetch.md`, `docs/decisions/0012-PLAN-repair-duckduckgo-news-video-and-fetch.md`, `internal/engine/news_shape_test.go`, `internal/engine/video_shape_test.go`, `internal/engine/fetch_shape_test.go`, `internal/handler/search/fetch_result_shape_test.go`, `internal/handler/search/news_endpoint_test.go`, `internal/handler/media/video_page_test.go`, `internal/handler/media/image_endpoint_test.go`, and `docs/README.md`.

## Rollback

Revert the implementation commit. `NewsSearch` and `VideoSearch` call `BingWebSearch` again. news.js goes back to the `fmt.Sprintf` URL with no `noamp` and no `df`. Video URLs go back to the empty `url` field. Empty HTML goes back to a blank `Content`. `SearchMetadata` loses `endpoint`. No migration, no config key, no cache format change. VQD entries cached under a rewritten query expire under the existing VQD TTL. Nothing in `scraping.db` has to be deleted.

## Completion Criteria

1. Each new test named in Phases 2 through 5 passed on the live tree. On `/tmp/ddg-0012-plant`, before the production edit, each first-group test failed an assertion and each second-group test failed to compile. The execution record quotes those lines.
2. `TestNewsSearch/fallback_google_news`, `TestVideoSearch/fallback_google_video`, `TestWebSearch/fallback_bing`, and `TestFetchContent_Success` still pass.
3. The four single-package `go test` commands, `check_records.py --check-all`, and `markdownlint-cli2` exit 0.
4. No product edit landed before the MADR was accepted and the owner said to proceed.

## Execution record

Populated as phases run. A deviation gets a dated entry here and, when it changes a fact in the MADR, an amendment on the MADR before the next phase.

### 2026-11-20 — owner accepted the MADR and said to proceed

Grounding re-read at `bb5b91c` (after the 0011 push). `NewsSearch` still called `BingWebSearch` at `news.go:29-34`, the news URL was still the `fmt.Sprintf` at `news.go:43`, the excerpt copy was still `news.go:99`, `VideoSearch` still called `BingWebSearch` at `video.go:28-33`, the video struct still decoded `url` at `video.go:64-73`, and the result URL was still `r.URL` at `video.go:86`. Fetch selector, marker, and cell lines matched the plan. `SearchMetadata` was still the three fields at `models.go:24-29`.

Plant on `/tmp/ddg-0012-plant`, a local clone of `bb5b91c`, before the production edit. First-group failures:

```text
TestNewsSearch/query_shape/keyword: params noamp="" p="-1" l="us-en" o="json"
TestNewsSearch/empty_query/today: Do called for "today": https://duckduckgo.com?q=today
TestNewsSearch/excerpt_plain: description "Tom&#x27;s <b>Fed</b>"
TestNewsSearch/no_rows_skips_bing: bing host "www.bing.com"
TestVideoSearch/content_url: url ""
TestVideoSearch/embed_url_fallback: URL empty
TestVideoSearch/no_rows_skips_bing: bing host "www.bing.com"
TestFetchContent_EmptyRoot: content ""
TestFetchContent_SkipsEmptyArticle: content ""
TestExtractStructuredText/two_cells: trailing separator: "Left | Right |"
TestFetchResultContents_FetchError: content ""
```

`TestExtractStructuredText/one_cell` first passed, because `TrimSpace` drops the trailing space of ` | ` and the old text was `"```\npackage main\n```\n |"`. The assertion is now that the result contains no `|`. That form failed on the plant with that same text.

`TestNewsSearch/rows_skip_bing`, `TestVideoPageURL`, and `TestFetchResultContents_EmptyMarker` passed on the plant. Bing runs only after `runProviders` returns an error, the media handler already copies `URL`, and a non-empty marker was already stored. Those three stay as regressions. The failing pins are `no_rows_skips_bing`, `content_url`, and `TestFetchResultContents_FetchError`.

Compile failures on the plant:

```text
undefined: newsToolDesc
r.Data.Metadata.Endpoint undefined (type *models.SearchMetadata has no field or method Endpoint)
```

The same four packages on the live tree then exited 0: `internal/engine` 13.899s, `internal/handler/search` 0.293s, `internal/handler/media` 0.271s, `internal/models` 0.216s. `gofmt -l` on the edited Go files printed nothing.

Deviation, same date. The new tests sit in sibling files in those packages: `news_shape_test.go`, `video_shape_test.go`, `fetch_shape_test.go`, `fetch_result_shape_test.go`, `news_endpoint_test.go`, `video_page_test.go`, and `image_endpoint_test.go`. They are part of this plan's file list.

### 2026-11-20 — gocognit on the first pre-add

`make pre-add-check` exited 2. golangci-lint reported two functions over its minimum of 50. Both were introduced by this plan. Neither was over the minimum at `bb5b91c`.

```text
internal/engine/fetch.go:65:1: cognitive complexity 51 of func `(*SearchEngine).FetchContent` is high (> 50) (gocognit)
internal/engine/news_shape_test.go:12:1: cognitive complexity 79 of func `runNewsShapeTests` is high (> 50) (gocognit)
```

The blank-content branch inside `FetchContent` caused the first. Nested `t.Run` closures inside `runNewsShapeTests` caused the second.

Resolution, same date. `extractStructuredText` returns `[No readable content extracted]` when its trimmed text is empty.
`FetchContent` still calls that function before the 15,000-rune truncation and does not call Playwright on that path.
`runNewsShapeTests` only dispatches. The query case body is `checkNewsQueryShape`.
`empty_query`, `excerpt_plain`, `no_rows_skips_bing`, and `rows_skip_bing` are package-level functions.
Subtest names are unchanged. No assertion was dropped and no `nolint` directive was added.

Re-run on the session host after that split.
`gofmt -l` on `internal/engine/fetch.go` and `internal/engine/news_shape_test.go` printed nothing.
`golangci-lint run -c .golangci.yml ./...` reported `0 issues`.

```text
ok github.com/maccavelli/mcp-server-duckduckgo/internal/engine 14.410s
ok github.com/maccavelli/mcp-server-duckduckgo/internal/handler/search 0.633s
ok github.com/maccavelli/mcp-server-duckduckgo/internal/handler/media 0.232s
ok github.com/maccavelli/mcp-server-duckduckgo/internal/models 0.146s
```

`python3 scripts/check_records.py --check-all` exited 0.
`npx --yes markdownlint-cli2@0.23.2` exited 0: 0 issues across 46 files.
`make pre-add-check` on the 17 Go files exited 0:
`go-precheck: 17 file(s) clean (gofmt, golangci-lint, go vet, go test, govulncheck).`

`git status --short` on `main` at `bb5b91c` lists only paths named in phase 6.
`internal/handler/media/tools_test.go` is clean.
`TestVideoPageURL` is in `video_page_test.go`.
