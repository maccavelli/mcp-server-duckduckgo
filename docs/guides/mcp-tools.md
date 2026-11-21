# MCP tools

Task index: [docs/README.md](../README.md).

The server registers seven tools. Search tools share `query`,
`max_results`, and `format`. `max_results` defaults to 5 when it is
omitted or not positive. `format` is `hybrid`, `json`, or `markdown`.
The default is `json`, or the value of `DDG_DEFAULT_FORMAT` when that
variable is set.

`json` returns the structured result. `markdown` returns rendered Markdown as text content. `hybrid` returns the structured result and also fills `results_md`.

## search_web

Description, from `internal/handler/search/tools.go`:

```text
[DIRECTIVE: General Internet Discovery] Comprehensive retrieval utilizing the DuckDuckGo index to extract current, real-time web results and engine queries. Keywords: web, internet, websites, online, general-search, urls, text, browse
```

| Field | Required | Meaning |
| :--- | :--- | :--- |
| `query` | yes | Search query. |
| `max_results` | no | Maximum results. Default 5. |
| `format` | no | `hybrid`, `json`, or `markdown`. |
| `fetch_content` | no | When true, fetch each result URL. Default false. |

Registered in `internal/handler/search/tools.go`.

## search_news

Description, from `internal/handler/search/tools.go`:

```text
[DIRECTIVE: Live Events and Journalism] Preferred choice for breaking developments, real-time updates, and current timeline information. Pass a short keyword query. The phrases today, right now, this week, and this month set freshness and are removed from the query sent to the news endpoint. Keywords: news, press, breaking, events, articles, temporal, journalism, headlines
```

`search_news` sends the keyword query to DuckDuckGo news with `noamp=1`.
`today` and `right now` set freshness to the past day (`df=d`).
`this week` sets the past week (`df=w`). `this month` sets the past month (`df=m`).
A query that is only one of those phrases returns the error `news query is empty` and does not call the network.
News and video misses are errors. They are not replaced with Bing web results.

The fields are the same as `search_web`, including `fetch_content`.

## search_books

Description, from `internal/handler/search/tools.go`:

```text
[DIRECTIVE: Book Catalog Lookup] Open Library work search for title, author, and publication year. Keywords: books, works, authors, reading, open-library
```

The fields are the same as `search_web`, including `fetch_content`.

`BookSearch` in `internal/engine/books.go` sends one `GET` to
`https://openlibrary.org/search.json`. A blank query returns an error and
does not call the network. A non-positive `max_results` becomes 5. A value
above 50 becomes 50. A work with a title and a work key becomes a result
whose `source` is `Open Library`, whose `url` is the work URL, and whose
`author` and `date` come from the author list and `first_publish_year`.

## search_images

Description, from `internal/handler/media/tools.go`:

```text
[DIRECTIVE: Visual Asset Retrieval] High-priority visual discovery to locate UI/UX inspiration, technical diagrams, photographs, and logos. Keywords: images, pictures, graphics, photos, diagrams, visual, assets, layouts
```

| Field | Required | Meaning |
| :--- | :--- | :--- |
| `query` | yes | Search keywords. |
| `max_results` | no | Maximum results. Default 5. |
| `format` | no | `hybrid`, `json`, or `markdown`. |

There is no `fetch_content` field. Registered in `internal/handler/media/tools.go`.

## search_videos

Description, from `internal/handler/media/tools.go`:

```text
[DIRECTIVE: Multimedia and Demonstration Retrieval] Targeted search for tutorials, walkthroughs, multimedia reports, and movies. Keywords: videos, multimedia, tutorials, walkthroughs, movies, clips, streaming, media
```

The fields are the same as `search_images`.

## fetch_url

Description, from `internal/handler/search/tools.go`:

```text
[DIRECTIVE: Deep Content Retrieval] Fetches and extracts clean, readable text content from any URL using the hardened HTTP client with cookie persistence, UA rotation, and CAPTCHA detection. Use after search to read full page content. Keywords: fetch, read, extract, scrape, content, page, article, url
```

| Field | Required | Meaning |
| :--- | :--- | :--- |
| `url` | yes | The URL to fetch. |

`fetch_url` is the page-fetch tool. `read_url` is not a registered name.

## get_internal_logs

Registered by `internal/handler/system/tools.go` through `mcplib` `v1.6.0`
`RegisterDiagnosticTool`, with the domain `duckduckgo`. The description
`mcplib` builds starts with `[DIRECTIVE: DUCKDUCKGO TELEMETRY]` and says
the tool retrieves diagnostic logs from that server's memory buffer.

| Field | Required | Meaning |
| :--- | :--- | :--- |
| `session_id` | no | When set, keep only log lines that contain this string. |
| `max_lines` | no | Number of recent lines. Omitted or not positive means 25. Values above 1000 become 1000. |

An empty buffer returns the text `[NO LOGS AVAILABLE]`.

## Names that are not tools

`search_suggest` is not registered.
