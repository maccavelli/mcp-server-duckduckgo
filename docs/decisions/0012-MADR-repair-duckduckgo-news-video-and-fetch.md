---
status: accepted
date: 2026-11-20
---
# Repair DuckDuckGo news queries, video URLs, and empty page fetches

## Context and Problem Statement

`search_news`, `search_videos`, `search_images`, `search_web`, and `fetch_url` were called through the live `ddg-search` server on 2026-11-20. That process logged `version` `1.2.0`. The engine files behind those calls match this tree at `05474ff` (no diff against tag `v1.2.0` for `internal/engine/{news,web,fetch,image,video,bing,bing_web,google,runner}.go`).

Three symptoms showed up, and the source explains each one.

News queries are sent to `https://duckduckgo.com/news.js` as the caller typed them. The request sets `l=us-en`, `o=json`, and `p=-1`. It does not set `noamp` or `df`. A keyword query and a full question therefore hit the same URL shape. Excerpts are copied through with highlight tags and HTML entities still in the string. When that call returns no rows, `NewsSearch` calls `BingWebSearch`, which is `https://www.bing.com/search` and labels every hit `source` `Bing`.

Video hits come back with a title, a duration, and a publisher, and with `page_url` empty. The decoder reads JSON field `url`. The current DuckDuckGo video payload puts the watch URL in `content`.

`fetch_url` returns article text for server-rendered pages and returns a title plus an empty body for a page whose HTML is an empty `#root`. That empty body is a successful tool result. The empty-content marker in `FetchContent` is not used, because a node was selected.

The live calls, and what they show:

* `search_web` for `golang net/http server example` returned five hits, each `source` `DuckDuckGo`, with direct `https` URLs. A POST to `https://html.duckduckgo.com/html/` for `golang http server` from this host returned HTTP 200, ten `result__a` hrefs, and zero `uddg` tokens. Those hrefs were the article URLs.
* `search_news` for `Federal Reserve interest rate` returned five hits with publisher names, `image_url`, and RFC3339 dates (the unix-seconds branch in `ddgNewsSearch`). Descriptions still contained `<b>` around matched tokens. One hit was dated `2024-09-18`. One URL was `www.msn.com`.
* `search_news` for `what is the federal reserve doing with interest rates right now` returned five hits. Four were The Motley Fool. Four URLs were `www.msn.com`. Descriptions contained `<b>` and `&#x27;`. Highlighted tokens included `The` and `rate`.
* `fetch_url` of the Yahoo Finance URL from the first news call returned the article body. `fetch_url` of `https://www.fool.com/investing/2026/11/15/how-will-the-market-react-to-a-potential-fed-rate/` returned the article body. `fetch_url` of the Motley Fool MSN URL returned the heading `MSN`, the source URL, and no body text.
* A direct GET of that same MSN URL from this host returned HTTP 200, `content-type: text/html; charset=utf-8`, `<title>MSN</title>`, one `id="root"`, no `article`, no `main`, and three characters of visible text: `MSN`.
* `fetch_url` of `https://example.com`, `https://gobyexample.com/http-servers`, and `https://en.wikipedia.org/wiki/Go_(programming_language)` returned body text. The Go by Example page and the Wikipedia page were split with ` | ` between table cells. The Wikipedia body ended with `[Content truncated at 15,000 characters]`.
* `search_images` for `hubble space telescope repair` returned three hits with real titles, publisher page URLs, full image URLs, thumbnails on `tse*.mm.bing.net`, and `source` `Bing`. `BingImageSearch` in this tree sets `source` to `Bing Images` and `title` to `Image result`, and it does not set a page URL. These hits are the `i.js` shape, with the payload's own `source` string.
* `search_videos` for `origami crane folding tutorial` returned three hits. Each had a title, `duration` (`3:56`, `0:43`, `0:49`), and `publisher` `YouTube`. Each `page_url` was `""`. No `media_url` was present. `BingWebSearch` does not set `duration` or `publisher`. These hits are the `v.js` shape with the watch URL dropped.

`news.js`, `v.js`, and `i.js` answered HTTP 403 to a cookie-jar client on this host that was not the server's own client. The parameter lists below are taken from the engine source and from the current `ddgs` clients fetched on 2026-11-20 (`raw.githubusercontent.com/deedy5/ddgs/main/ddgs/engines/duckduckgo_news.py`, `duckduckgo_videos.py`, and `duckduckgo_images.py`). Those clients set `provider = "bing"` for all three verticals. A 403 from the probe client is not evidence that the server's client is rejected. The server returned vertical-shaped rows in the same hour.

The news client builds this payload: `l`, `o=json`, `noamp=1`, `q`, `vqd`, `p` (`1` on, `-1` moderate, `-2` off), plus `df` when the caller passes a timelimit (`d`, `w`, `m`, `y`), plus `s` on later pages. This server's `ddgNewsSearch` sends `q`, `vqd`, `l=us-en`, `o=json`, `p=-1`. `p=-1` is that client's moderate setting. `noamp` and `df` are absent. There is no MCP field for freshness.

The video client reads the watch URL from JSON field `content`. Sibling fields include `description`, `duration`, `embed_url`, `images`, `provider`, `published`, `publisher`, and `title`. There is no `url` key in that map. A published example value of `content` is `https://www.youtube.com/watch?v=...`, and the example `provider` is `Bing`. `ddgVideoSearch` decodes `url`, `description`, `duration`, `publisher`, and `published`. `internal/handler/media/tools.go` copies `URL` into both `page_url` and, when `image_url` is empty, `media_url`. An empty `url` therefore yields an empty `page_url` and no `media_url`. The description is not copied onto `MediaResult20`.

The image client reads `title`, `image`, `thumbnail`, `url`, and `source`. Thumbnails in the published video example use `tse*.mm.bing.net`, the same host as the live image thumbnails. `ddgImageSearch` copies `source` through. The live `source` value `Bing` matches that field. It does not match `providerBingImages` (`Bing Images`).

`FetchContent` removes noise, then keeps the first node that matches `article`, `[role='main']`, `main`, a list of content classes, `#content`, or `.content`. A later selector is not tried when the first match has no text. If none match, a text-density walk may select a node, and otherwise the `body`. The marker `[No readable content extracted]` is returned only when no node remains. `extractStructuredText` appends ` | ` for every `td` and `th`. Playwright runs only for HTTP 403 or for the DuckDuckGo challenge strings in `detectDDGChallenge`. A 200 response whose text is empty does not take that path. `fetchResultContents` leaves `Content` empty when `FetchContent` returns an error or an empty string.

`parseGoogleWebResults` is what `GoogleSearch` uses for `tbm=nws` and `tbm=vid` as well as for web. It keeps `div.g` rows with an `h3`, an `http` href, and `.VwiC3b, .y67o5b, .MUF9Of`. `TestNewsSearch/fallback_google_news` feeds a synthetic `div.g` document. An empty slice from that parser is a nil error. `runProviders` then returns `no results found` once every provider has contributed nothing, and `recordProviderResult` counts an empty slice as a failure. Three consecutive failures make `IsHealthy` false until `HealthTTL` (one hour). `NewsSearch` and `VideoSearch` then call `BingWebSearch`. `WebSearch` does the same. This session's web and news calls returned DuckDuckGo-shaped rows, so that Bing web hop was not taken on those calls.

## Decision Drivers

* A news call should send the parameters the news endpoint uses for a news query: the query string, `noamp`, and a freshness value when the caller asked for one.
* A video hit should carry the watch URL the payload already contains.
* A fetch of a page with no readable text should say so in the tool result.
* A failed news or video call should stay a failed news or video call. `BingWebSearch` is a web SERP.
* Publisher and upstream strings that the payload already sends should remain visible. The endpoint that was called should be visible as well.
* The repair stays on the existing HTTP client, `models.SearchResult`, and the seven registered tools. No new module.

## Considered Options

* Repair DuckDuckGo news, video, and fetch in place, and stop the news and video Bing web hop
* Send news and video to Bing's news and video pages instead of `news.js` and `v.js`
* Strip excerpt HTML only, and leave the request, the video decoder, the fetch marker, and the Bing web hop as they are

## Decision Outcome

Chosen option: "Repair DuckDuckGo news, video, and fetch in place, and stop the news and video Bing web hop", because the live news and video rows already come from `news.js` and `v.js`, the missing pieces are on the request and the decoder, and the empty MSN body is an empty HTML document reported as success. Replacing those calls with Bing's own pages would drop the endpoint the server is supposed to call. Stripping excerpt tags would leave the question-shaped news request, the empty video URL, and the silent empty fetch in place.

This record is `proposed`. The state table rows are `open` until the owner rules on them. The recommended ruling is in each row.

### Confirmation

* A news unit test asserts the `news.js` query string contains `noamp=1` and contains `df` only for the recency cases the owner accepts.
* A news fixture whose `excerpt` contains `<b>` and `&#x27;` asserts a plain-text `Description`.
* A news test whose DuckDuckGo and Google providers return no rows asserts an error and no request to `bing.com`.
* A video fixture shaped like the `v.js` element map (`content` set, `url` absent) asserts `page_url` equals that `content` value.
* A fetch fixture whose document is `<title>MSN</title><body><div id="root"></div></body>` asserts the empty-content marker and a nil error.
* A fetch fixture with a one-cell layout table asserts the cell text without a pipe between every cell.
* The existing `TestFetchContent_Success` path still returns the article sentence.
* Each new assertion is shown to fail on a deliberately broken copy before it is kept. That experiment stays off the tree.

## Consequences

* Good, because a news request can carry `noamp=1` and a freshness `df` when the query asks for a time window, which is the news payload the current `ddgs` news client builds.
* Good, because video hits expose the watch URL that `v.js` already returns, so `fetch_url` has a URL to call.
* Good, because an empty document returns the empty-content marker instead of a blank body under a title.
* Good, because a news or video miss is an error from that vertical. Callers can see the miss.
* Neutral, because DuckDuckGo's own news, image, and video clients document `provider = "bing"`. Hits can still say `Bing` in the payload `source` or `provider` field after this repair. That string is the upstream index, and it remains visible.
* Neutral, because `search_web` keeps its Bing web hop. The live web call and the HTML probe both returned DuckDuckGo HTML results with direct hrefs.
* Bad, because callers who relied on `NewsSearch` or `VideoSearch` to fill a failure with Bing web links will receive an error instead.
* Bad, because `noamp=1` was not A/B tested from this host. `news.js` returned 403 to the probe client. The flag is still the one the news client sends. Whether it changes the share of `msn.com` URLs is unmeasured.
* Bad, because a 200 page whose article text is created by JavaScript still has nothing to extract. The marker reports that. It does not run a browser.

## Pros and Cons of the Options

### Repair DuckDuckGo news, video, and fetch in place, and stop the news and video Bing web hop

* Good, because the live news, image, and video calls already reached `news.js`, `i.js`, and `v.js`. The gaps are the query string, the JSON keys, and the empty-body result.
* Good, because `BingWebSearch` stays available to `WebSearch`, where a web SERP is the same kind of result.
* Neutral, because image search already returns `i.js` rows. This option does not replace `BingImageSearch`. It stops treating the payload `source` `Bing` as proof that `BingImageSearch` ran.
* Bad, because freshness defaults, question rewriting, and Playwright-on-empty are still owner choices. They are the open rows below.

### Send news and video to Bing's news and video pages instead of `news.js` and `v.js`

* Good, because the upstream index is already Bing, so a first-party Bing news or video page would match the index the user is seeing.
* Bad, because the server's job for these tools is the DuckDuckGo endpoint. The live news and video calls that succeeded were those endpoints.
* Bad, because this tree has no Bing news or Bing video parser. `BingWebSearch` and `BingImageSearch` are different pages, and the image parser's result shape (`Image result`, `Bing Images`, no page URL) is the wrong shape for the hits already returned.

### Strip excerpt HTML only, and leave the request, the video decoder, the fetch marker, and the Bing web hop as they are

* Good, because descriptions would stop showing `<b>` and `&#x27;`, which the live news JSON contained.
* Bad, because the question-shaped query, the missing `df` and `noamp`, the empty video `page_url`, and the blank MSN body would remain.
* Bad, because a news miss would still become a Bing web SERP labeled `Bing`.

## State table

| State | Ruling | Test that pins it |
| --- | --- | --- |
| News `q` is a keyword phrase and names no time window | decided: send `q` unchanged, always send `noamp=1`, omit `df`. | Request URL contains `noamp=1` and the escaped query, and does not contain `df=`. |
| News query contains a recency phrase (`today`, `right now`, `this week`, `this month`) | decided: set `df` to `d`, `d`, `w`, or `m` respectively, and remove that phrase from `q`. | One fixture per phrase asserts `df` and the remaining query. |
| News query is only a recency phrase | decided: return an error and do not call the network. | Fixture query `today` asserts an error and zero requests. |
| News query is a leading question and has no recency phrase | decided: send the words unchanged. State in the `search_news` description that a short keyword query is the news query. An automatic rewrite is unmeasured. The live question returned four Motley Fool hits and four `msn.com` URLs; the keyword query did not. | Fixture asserts `q` equals the caller string, and the tool description names a keyword query. |
| News `excerpt` contains `<b>` or an HTML entity | decided: unescape, drop tags, then truncate to `MaxSnippetLength`. | Fixture excerpt `Tom&#x27;s <b>Fed</b>` asserts description `Tom's Fed`. |
| `news.js` returns one or more rows | decided: return those rows and do not call Bing. | Mock `news.js` 200 with one row; assert the row and zero requests whose host is `bing.com`. |
| DuckDuckGo news and Google news both return no rows | decided: return the existing no-results error. Do not call `BingWebSearch`. | Both providers empty; assert the error and zero `bing.com` requests. |
| `v.js` item has `content` and no `url` | decided: set the result URL from `content`. Copy `embed_url` only as a fallback when `content` is empty. Keep `publisher` and `duration`. | Fixture with `content` `https://www.youtube.com/watch?v=abc` and `publisher` `YouTube` asserts `page_url` and `media_url` are that watch URL. |
| `i.js` item has `source` `Bing` | decided: keep that string on the hit. Add the endpoint name `DuckDuckGo` on `SearchMetadata` for news, images, and videos. | Image fixture asserts hit `source` `Bing` and metadata endpoint `DuckDuckGo`. |
| Fetch HTTP 200 HTML whose visible text is empty (the MSN document) | decided: return nil error and content `[No readable content extracted]`. Do not launch Playwright for this record. | Fixture `<title>MSN</title><body><div id="root"></div></body>` asserts that marker. |
| Fetch HTTP 200 HTML with article text | decided: keep today's success path. Skip a content selector whose trimmed text is empty and try the next selector. | Existing success fixture, plus a document whose first `article` is empty and whose `main` holds the sentence. |
| Fetch HTTP 200 HTML that uses a one-cell layout table | decided: emit a cell separator only for a row that has two or more cells. | Go by Example-shaped fixture asserts the code line with no cell separator on either side. |
| `fetch_content` on a search hit whose fetch returns the empty marker or an error | decided: store the marker in `Content`. On error, store the error string in `Content` and keep the hit. | Handler test with a fetch error asserts `Content` holds that string. |
| Google `tbm=nws` or `tbm=vid` document does not match `div.g` | out of scope. An empty parse contributes no rows. This record does not add a news or video DOM parser for Google. | None. |
| `search_web` primaries fail | out of scope. `WebSearch` keeps `BingWebSearch`. The live web call returned DuckDuckGo HTML hits. | Existing `TestWebSearch/fallback_bing` stays. |
| `i.js` fails and `BingImageSearch` runs | out of scope. That parser already labels `Bing Images`. The live image call was `i.js`. | Existing image fallback test stays. |
| Region `l=us-en`, page size, and the 15,000-rune cap | out of scope. No probe showed a wrong-region miss. The cap fired on the Wikipedia fetch and is labeled. | None. |
| HTTP 403 or DuckDuckGo challenge during fetch | out of scope. Playwright remains the path for that case only. | Existing challenge coverage stays. |

## More Information

Live tool calls used `ddg-search:search_news`, `ddg-search:search_web`, `ddg-search:search_images`, `ddg-search:search_videos`, and `ddg-search:fetch_url` with `format` `json` where the tool accepts it. `get_internal_logs` showed the server process started `2026-11-19T17:25:56-06:00` as `mcp-server-duckduckgo` `1.2.0`, and showed `VQD cache miss` lines for the news and image queries above. It did not show a Bing request line. Provider failures are logged at warn, and this buffer's lines for those calls were info.

`internal/engine/news.go` `ddgNewsSearch` builds the `news.js` URL. `NewsSearch` calls `BingWebSearch` when `runProviders` returns an error. `internal/engine/video.go` does the same hop. `internal/engine/google.go` `parseGoogleWebResults` is the parser for `tbm` `nws` and `vid`. `internal/engine/fetch.go` `FetchContent` and `extractStructuredText` implement the body extraction. `internal/handler/media/tools.go` copies video `URL` into `page_url` and `media_url`. `internal/handler/search/tools.go` `fetchResultContents` ignores an empty `Content`.

The HTML probe and the MSN probe ran on this host on 2026-11-20 via `/tmp/ddg-probe.py` (removed after the run). Homepage VQD patterns matched four times. The HTML endpoint returned ten direct result links. The MSN document's visible text was `MSN`.

The owner accepted this record on 2026-11-20 and said to proceed. Each row that was open is decided as the ruling in that row. [0012-PLAN-repair-duckduckgo-news-video-and-fetch.md](0012-PLAN-repair-duckduckgo-news-video-and-fetch.md) is approved and implements those rulings.
