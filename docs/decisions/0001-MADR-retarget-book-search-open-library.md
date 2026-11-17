---
status: accepted
date: 2026-11-16
---
# Retarget book search from Anna's Archive to Open Library Search API

## Context and Problem Statement

`search_books` (`SearchEngine.BookSearch` in `internal/engine/mirror.go`) scrapes Anna's Archive HTML at `https://annas-archive.{gd,gl,pk}/search`. That path is no longer usable from this server's HTTP client.

Live reproduction on 2026-11-16 (UTC evening of 2026-11-17 on the origin): MagicTools `ddg-search:search_books` for `python programming` returned `book search failed across all mirrors: mirror returned status 403`. Direct GETs to `/search` on `.gd`, `.gl`, and `.pk` answered `server: ddos-guard`, redirected to `?check=1`, then 403 with a DDoS-Guard JavaScript challenge page. Homepages, `/account`, `/datasets`, `/llms.txt`, and `/dyn/torrents.json` still returned 200. Homepage cookies did not unlock `/search`. `annas-archive.org` and `.se` did not resolve; `.li` is a parked page.

Anna's Archive publishes no public search API. Their `llms.txt` directs machines at bulk torrents and a member-only download-by-MD5 JSON endpoint. Signed-in members skip the search challenge; anonymous HTML clients do not. The current scraper also fans out three concurrent mirror GETs, which is the wrong pattern against that gate.

The owner asked to retarget book search at a fast public endpoint rather than authenticate to Anna's Archive, solve the JS challenge, or host a local metadata dump.

The engine already speaks JSON over `net/http` (see `NewsSearch`). `models.SearchResult` already has `Title`, `URL`, `Description`, `Author`, `Publisher`, `Date`, `Info`, `Thumbnail`, and `Source`. The MCP tool name `search_books` and `BookSearch` signature stay.

Which public catalog should `BookSearch` call?

## Decision Drivers

* No API key and no membership secret in process config or git.
* Latency well inside `config.DefaultTimeout` (15s), preferably sub-second for a five-result query.
* A book catalog, not a paper/DOI index, matching the tool name `search_books`.
* Stable JSON that maps onto `models.SearchResult` without HTML scraping.
* Polite, documented client behaviour (identifying User-Agent, bounded `limit`, no concurrent stampede).
* Fits the existing hardened HTTP client; no new runtime dependency.

## Considered Options

* Open Library Search API (`GET https://openlibrary.org/search.json`)
* Google Books API (`GET https://www.googleapis.com/books/v1/volumes`) without a configured key
* OpenAlex Works API
* Crossref Works API
* Keep scraping Anna's Archive (member session, Playwright JS-check, or bulk dump)

## Decision Outcome

Chosen option: "Open Library Search API (`GET https://openlibrary.org/search.json`)", because official Books API docs rank that Search API as the preferred multi-book entrypoint, a live probe from this host returned HTTP 200 in 0.60s wall time (Solr 62ms, total server 77ms via `server-timing`) for `q=python+programming&limit=5` with an explicit `fields` list, required no key, and returned book works with title, authors, year, ISBN, and a work key.

Google Books from the same host returned HTTP 429 with `quota_limit_value: "0"` on the anonymous default project. OpenAlex and Crossref answered quickly but returned conference papers and book-chapters, not a book catalog. Anna's Archive search remains 403 for this client.

### Consequences

* Good, because `search_books` uses the documented public JSON Search API instead of HTML. Open Library's usage guidelines instruct clients to use API endpoints and not scrape HTML pages.
* Good, because restricting `fields` is the documented performance contract after the 2025-01-21 default-field reduction. The client asks for `number_of_pages_median` (absent from the OpenAPI default list) and does not ask for `isbn` or `publisher`: those work-level fields are edition bags, and a five-hit "lord of the rings" response was 38,206 bytes with them and 1,803 bytes without.
* Good, because result URLs are `https://openlibrary.org` + work key, which `fetch_url` / `fetch_content` can retrieve as ordinary HTML.
* Neutral, because Open Library rate-limits unidentified clients to 1 request per second and identified clients (User-Agent plus contact email) to 3 requests per second. Book search is a single GET per tool call, so this matches MCP use.
* Neutral, because the corpus is Open Library's public catalog: works, editions, ISBNs, and borrow/preview links where the Internet Archive has them.
* Neutral, because Search returns Works by default (FRBR: one work, many editions). Callers see work-level title/author/year plus selected edition fields on the work doc. One row per printing would be a later decision on `/search/editions.json` or the nested `editions` field.
* Bad, because callers who depended on Anna's Archive MD5 detail URLs will see different URLs and a different collection.
* Bad, because Open Library's Solr field set is documented as not strictly stable; the client must request a named field list and tolerate missing optional fields.

### Confirmation

* Unit tests feed a fixture shaped like the 2026-11-16 `search.json` body (`numFound`, `num_found`, `docs`, `offset: null`, `key` values `/works/OL…W`) into `BookSearch` through an `http.Client` mock and assert mapped `SearchResult` fields, including a `/works/` URL.
* Those tests fail on truncated JSON, non-200, and empty `docs` before they are accepted as gates.
* After the change, `internal/engine` contains no `annas-archive` host construction.
* A live GET of the same URL used in the probe still returns 200 with `docs` when run by hand; that live call is not a CI gate.

## Pros and Cons of the Options

### Open Library Search API (`GET https://openlibrary.org/search.json`)

Official Book APIs page (edited 2025-05-06) ranks five book surfaces and names Search first: "If you are looking for a single API to meet all of your book needs, it's likely the Book Search API." Search docs (edited 2025-05-07) give the URL `https://openlibrary.org/search.json` and describe it as able to return data for multiple books in one request.

Selection probe 2026-11-16 (a wider field list than the one this decision maps), `q=python%20programming&limit=5&fields=key,title,author_name,first_publish_year,publisher,isbn,subject,number_of_pages_median,language,cover_i,edition_count`: HTTP 200, 2082 bytes, `time_total=0.596s`, `server-timing: total;dur=77.4, solr;dur=62.4`. First doc: title `Core Python Programming`, key `/works/OL30906747W`, author `R. Nageswara Rao`, year 2016. Same query without `fields`: HTTP 200, Solr 490ms. Title/author query `title=Clean+Code&author=Martin`: HTTP 200, 0.39s, first doc Robert C. Martin, 2008.

* Good, because it is a book catalog with title, author, year, ISBN, cover id, and a work URL, served as JSON over a single HTTPS GET with no API key.
* Good, because official conduct rules tell clients to use `search.json` for batch results rather than hundreds of single-book GETs, which matches one MCP tool call returning `limit` hits.
* Good, because it is free of keys and of DDoS-Guard on this host.
* Neutral, because polite use is 1 request per second unidentified, or 3 per second with `User-Agent: <app> (<email>)`.
* Bad, because optional fields may be absent and the Solr schema is not a compatibility contract.

### Google Books API (`GET https://www.googleapis.com/books/v1/volumes`) without a configured key

Probe 2026-11-16: HTTP 429, `RESOURCE_EXHAUSTED`, quota metric `Queries per day` for `books.googleapis.com`, `quota_limit_value: "0"` on consumer `project_number:624717413613` (the unauthenticated default).

* Good, because with a billed/key-bearing project it is a fast, familiar book index and maps cleanly onto `SearchResult`.
* Bad, because without a configured key this host is already over quota (limit value 0). It is not a zero-config public endpoint.
* Bad, because adding a Google API key contradicts the "no key" driver and the server's zero-configuration stance in the README.

### OpenAlex Works API

Probe 2026-11-16: HTTP 200, 0.66s, first result type `conference-paper` ("Python Programming Language.", Guido van Rossum). Generous `x-ratelimit-limit: 1000`.

* Good, because it is fast, keyless (polite pool prefers a `mailto`), and JSON.
* Bad, because the corpus is scholarly works, not a book catalog, so `search_books` would start returning papers.

### Crossref Works API

Probe 2026-11-16: HTTP 200, 0.64s, first result type `book-chapter` with a DOI URL. Headers: `x-rate-limit-limit: 1`, `x-rate-limit-interval: 1s`, `x-api-pool: public-array`.

* Good, because it is fast, keyless JSON with DOI landing URLs.
* Neutral, because 1 request per second matches Open Library's unidentified pool.
* Bad, because results are bibliographic works (chapters, articles) rather than library book records with covers and ISBNs as the primary object.

### Keep scraping Anna's Archive (member session, Playwright JS-check, or bulk dump)

* Good, because the existing HTML parser still matches how third-party clients parse AA cards (`a.js-vim-focus[href^='/md5/']`).
* Neutral, because a member `POST /account/` with field `key` is how AA currently exempts search, and `/dyn/api/fast_download.json` is their only stable JSON API.
* Bad, because the owner asked for a fast public endpoint, not a membership secret.
* Bad, because Playwright-through-DDoS-Guard fights the protection AA's `llms.txt` asks machines not to break, and the current Turnstile detector would not recognise DDoS-Guard.
* Bad, because `aa_derived_mirror_metadata` / ElasticSearch dumps are the official machine search path and are far too large for this MCP process.

## More Information

Implementation plan: [0001-PLAN-retarget-book-search-open-library.md](0001-PLAN-retarget-book-search-open-library.md). Current code: `internal/engine/mirror.go`, tests in `internal/engine/mirror_test.go`, MCP registration in `internal/handler/search/tools.go` (`Type: "books"`).

Official sources read 2026-11-16 (page "last edited" dates are Open Library's):

* API hub / usage / rate limits: <https://openlibrary.org/developers/api> (edited 2026-11-16)
* Search API: <https://openlibrary.org/dev/docs/api/search> (edited 2025-05-07)
* Search HowTo (Solr query language): <https://openlibrary.org/search/howto> (edited 2026-05-14)
* Books API (ranking of book surfaces): <https://openlibrary.org/dev/docs/api/books> (edited 2025-05-06)
* RESTful API (content negotiation, Query API, history, login, status codes): <https://openlibrary.org/dev/docs/restful_api> (edited 2023-09-01)
* Legacy Infogami JSON API: <https://openlibrary.org/dev/docs/json_api> (edited 2022-09-05)
* Covers API: <https://openlibrary.org/dev/docs/api/covers> (edited 2024-08-13)
* Authors API: <https://openlibrary.org/dev/docs/api/authors> (edited 2024-09-03)
* Subjects / Lists / My Books / Read / Search Inside / Recent Changes / Dumps / Bulk Data pages under `/dev/docs/api/` and `/developers/dumps`
* OpenAPI 3.1.0 schema: `GET https://openlibrary.org/openapi.json` (interactive UI at <https://openlibrary.org/docs>)
* FastAPI migration status: <https://docs.openlibrary.org/projects/fastapi-migration.html> (status August 2026)
* Search.json default-fields change: <https://blog.openlibrary.org/2025/01/16/api-search-json-performance-tuning/> (deployed 2025-01-21)
* Solr work-search field universe (source, not a stability contract): <https://github.com/internetarchive/openlibrary/blob/master/openlibrary/plugins/worksearch/schemes/works.py>
* Official Python client: <https://github.com/internetarchive/openlibrary-client>

### Platform, protocols, and transports

Open Library is a public-good book catalog (Internet Archive). The developer hub states APIs exist in JSON, YAML, and RDF/XML to support human-centered book discovery and lookup. They are not a bulk data backend or high-traffic commercial infrastructure.

Two serving stacks share `https://openlibrary.org`:

* **FastAPI** — high-traffic JSON routes. Migration docs (August 2026): most high-traffic endpoints have moved; unmatched requests proxy to web.py. Interactive schema is OpenAPI 3.1.0 at `/openapi.json` (50 paths). `/search.json` is in that schema (`operationId` `search_json_search_json_get`, tag `search`). FastAPI validation failures return **422**. Cookie auth scheme in the schema is `APIKeyCookie` named `session`.
* **Infogami / web.py REST** — "nearly every item on Open Library has an API version." Work, edition, author, and ISBN HTML pages become structured data by appending `.json`, `.yml`, or `.rdf`, or by sending `Accept: application/json` / `Accept: application/rdf+xml`. YAML is documented on the hub and on work/edition pages as a URL suffix; the REST page's Accept-header examples are JSON and RDF. Optional `callback` wraps JSON as JSONP. History of an object: `?m=history`. These Infogami resource paths (`/works/{id}.json`, `/books/{id}.json`, `/isbn/{isbn}.json`, `/query.json`) are **absent** from `/openapi.json`.

Transport for `BookSearch` is a single **HTTPS GET** to `openlibrary.org`. There is no documented WebSocket, gRPC, or streaming search transport. Covers live on a **separate host**, `covers.openlibrary.org`, as JPEG (and optional `.json` metadata). Search-inside hits archive.org datanodes (`inside.php`), not `openlibrary.org`.

HTTP status codes documented on the REST page: 200, 400 (bad input, message in body), 403 (PUT/POST without permission), 404, 500. FastAPI adds 422 for schema validation.

### Conduct, identification, and rate limits

From the API hub:

* Do: make time-sensitive requests on behalf of human users; cache; identify the application with a User-Agent and email.
* Do not: scrape HTML or harvest HTML in bulk; distribute requests across multiple IPs; make hundreds of single-book requests (use `search.json` for batch results); use Open Library as a backend for high-traffic commercial services.
* Violations may result in aggressive rate limiting or blocking.
* Register use-cases at the hub's registration form. Contact `openlibrary@archive.org`.
* Example identified header: `User-Agent: MyLibraryApp (contact@example.org)`.
* Default (non-identified): **1 request per second**. Identified (User-Agent and email): **3 requests per second**.

Bulk access is monthly dumps (`/developers/dumps`), not the APIs. Editions dump ~9.2G, works ~2.9G, authors ~0.5G; catalog size stated as over 20 million editions and some 6 million authors (`/data`). Query API and REST docs both say those APIs must not be used for bulk download.

Covers API: do not crawl covers; display via `src` URLs on `covers.openlibrary.org`. Cover access by keys **other than Cover ID and OLID** is limited to **100 requests per IP per 5 minutes**, then 403. Bulk covers are tar dumps on archive.org.

### Official command surface (read-side)

Hub index, plus OpenAPI paths and REST docs. Grouped by purpose. Auth is none unless noted.

**Search (Solr, FastAPI, JSON).** Preferred book entrypoint.

* `GET /search.json` — works search. Query params documented on the wiki: `q` (Solr; see Search HowTo), `fields` (`*` is expensive; special value `availability` pulls archive.org availability for the first `ia` id), `sort` (`new`, `old`, `random`, `key`, and others in `works.py`), `lang` (ISO 639-1 display preference; does not exclude other languages — exclusion is Solr `language:fre`), `offset`/`limit` and `page`/`limit` (`page` starts at 1). OpenAPI additionally types `title`, `author`, `publisher`, `isbn`, `oclc`, `lccn`, `contributor`, `subject`, `place`, `person`, `time`, `has_fulltext`, `author_key`, facet arrays, `language` as three-letter codes, `query` as a JSON-encoded Solr object, `spellcheck_count` (default 3), and `editions.sort`. OpenAPI default `limit` is **100**. HTML search URLs convert to JSON by changing `/search?` to `/search.json?`.
* `GET /search/authors.json` — `q`; OpenAPI default `fields=*`, default `limit` 100.
* `GET /search/editions.json` — editions Solr index (OpenAPI; "full-text search across edition fields (title, ISBN, publisher, etc.)" and optional `work_key`). Default `limit` 100. Not the ranked preferred Book API.
* `GET /search/inside.json`, `/search/subjects.json`, `/search/lists.json`.
* Search HowTo fielded query language (passed in `q`): `title:`, `author:`, `subject:`, `place:`, `person:`, `language:` (three-letter; `spa`/`jpn`/`mul`/`und`), `publisher:`, `publish_year:[* TO 1800]`, `first_publish_year`, `title_suggest`, `ddc:`, `lcc:`, `subject_key:` (exact, normalised), negative `-subject_key:`, `ebook_access:` (`no_ebook` / `printdisabled` / `borrowable` / `public`).

**Works, editions, ISBN (Infogami REST, JSON/YAML/RDF via suffix or Accept).** One object per GET.

* Work: `GET /works/{OLID}.json` (example `/works/OL45804W.json`). Editions of a work: `/works/{OLID}/editions.json`. Ratings `/works/{OLID}/ratings.json`, bookshelves `/works/{OLID}/bookshelves.json`.
* Edition: `GET /books/{OLID}.json`.
* ISBN: `GET /isbn/{isbn}.json` redirects to the edition; ISBN 10 or 13.

**Legacy Books API** (`/api/books`, also FastAPI `/api/books.json`). Google Books Dynamic Links–inspired. Query: `bibkeys` (`ISBN:`, `OCLC:`, `LCCN:`, `OLID:`), `format=json|javascript` (default javascript), `callback`, `jscmd=viewapi|data|details` (default `viewapi`; `details` is the REST object and is documented as less stable than `data`). Books API page: "legacy endpoint and may be phased out"; prefer Search.

**Legacy Read / Partner API.** `GET /api/volumes/brief/{isbn|lccn|oclc|olid}/{id}.json` and a multi-id form. Hathi-compatible; JSONP; IP-aware lending. Hub lists it as legacy.

**Authors.** Search `/search/authors.json?q=`. Record `/authors/{OLID}.json|.rdf|.yml` (suffix on the key, not on the human-readable slug). Works `/authors/{OLID}/works.json?limit=&offset=` (docs example: 50 default; `limit=1000` used in an example). No batch full-author fetch; Solr `key:(/authors/A OR /authors/B)` returns a subset.

**Subjects (experimental).** `GET /subjects/{name}.json` with `details`, `ebooks`, `published_in`, `limit`, `offset`.

**Lists / My Books / Recent Changes.** Lists: GET/POST user lists, seeds, editions, subjects; search `/search/lists.json`; delete POST; update PUT not implemented. My Books (public reading log): `/people/{user}/books/{want-to-read|currently-reading|already-read}.json`. Recent changes: `/recentchanges.json` plus date and kind variants; experimental.

**Query API (Infogami objects, not Solr).** `GET /query.json?type=&authors=&title=`. Empty parameter value includes that property; `*=` all properties. Default **20** results, max **limit** **1000**. Alternate: `query=` JSON dictionary. Explicitly not for bulk.

**Legacy Infogami JSON API.** `GET /api/things`, `/api/get`, `/api/versions`. `/api/search` is **deprecated**. 60 second query timeout. Older keys `/b/`, `/a/` appear in examples; current REST keys are `/books/`, `/authors/`.

**Covers (separate host).** `https://covers.openlibrary.org/b/$key/$value-$size.jpg` with `key` in `{ISBN, OCLC, LCCN, OLID, ID}` (case-insensitive), size `S|M|L`. Default missing cover is a blank image; `?default=false` returns 404. Author photos: `/a/$key/$value-$size.jpg`. Cover metadata: append `.json` (example `/b/id/12547191.json`).

**Write / auth surfaces (not used by `search_books`).** Login: `POST /account/login` with JSON `access`/`secret` S3 keys; `Set-Cookie: session=...`. Save: PUT to a resource URL is documented as **localhost-only**. Lists create/delete and check-ins require the session cookie. Every account has S3 keys.

**Bulk.** Monthly TSV+JSON dumps on archive.org; OPDS/BookServer mentioned from the data hub. Official dump JSON schemas live in `openlibrary-client` `olclient/schemata`.

### Chosen endpoint contract: `GET /search.json`

Search docs: default return is **works**, not editions. A work is a collection of editions (Wizard of Oz example: one work `OL18417W`, 1029 editions). Nested `editions` on a work doc returns matching editions of that work (currently one displayed edition; `editions.sort` overrides relevance). `BookSearch` stays on the work list and does not request `editions` in this decision.

Response shape from the wiki:

```json
{ "start": 0, "num_found": 629, "docs": [ { ... } ] }
```

OpenAPI `SearchResponse` requires `numFound`, `start`, `numFoundExact`, `num_found`, `q`, `offset`; `docs` defaults to `[]`; `additionalProperties` true; each doc is an open object. Wiki examples use `num_found`; nested-editions examples use `numFound` / `numFoundExact`. Live bodies in the contract probe carried both counts, equal to each other. The mapper uses `docs` and does not branch on the counts.

Wiki sample doc fields: `cover_i`, `has_fulltext`, `edition_count`, `title`, `author_name[]`, `first_publish_year`, `key`, `ia[]`, `author_key[]`, `public_scan_b`. The wiki sample `key` is the bare id `OL27448W`. That shape is what `/search/authors.json` emits: current `openlibrary/fastapi/search.py` strips the path with `doc["key"].split("/")[-1]` and the source comment calls that out. `/search.json` does not. On 2026-11-16 all 1,005 distinct work keys across the search responses were `/works/OL…W`, including all 1,000 docs on the `limit=1000` page. The same work the wiki names is live as `/works/OL27448W`. `GET https://openlibrary.org/works/OL30906747W` returned 200 and redirected to the slugged HTML page. `GET https://openlibrary.org/OL30906747W` returned 404. A bare work id is prefixed with `/works/` before it is used as a URL. Any other shape is skipped.

Solr fields are described by `works.py` `all_fields` (including `key`, `title`, `subtitle`, `cover_i`, `author_name`, `publisher`, `isbn`, `language`, `first_publish_year`, `number_of_pages_median`, `edition_count`, `ebook_access`, …). Search docs: "The schema is not guaranteed to be stable, but most common fields (e.g. title, IA ids, etc) should be safe to depend on."

`fields` after 2025-01-21: developers "are strongly encouraged" to set `fields` to what they need; `fields=*` restores the previous full projection. OpenAPI default `fields` include `key`, `title`, `subtitle`, `author_name`, `cover_i`, `edition_count`, `first_publish_year`, `language`, and several lending/series ids. They omit `isbn`, `publisher`, and `number_of_pages_median`. `openlibrary/fastapi/models.py` `parse_comma_separated_list` accepts both a comma-separated value and repeated `fields` parameters. `Pagination.limit` defaults to 100, minimum 0, and `/search.json` does not set a maximum: `limit=1000` returned 1,000 docs.

The mapped field list is `key,title,subtitle,author_name,first_publish_year,number_of_pages_median,cover_i,edition_count`. `isbn` and `publisher` stay off that list. `language` stays off it too: `language[0]` on `/works/OL27448W` was `fre` in an 18-value list, so the first code is not the work's language.

Cover thumbnail for a search hit: Covers API `https://covers.openlibrary.org/b/id/{cover_i}-M.jpg` (Cover ID, which is in the un-rate-limited key class). The python-programming hit's Cover ID returned `image/jpeg` (20,820 bytes) after a redirect onto archive.org. This decision stores that URL on `SearchResult.Thumbnail` and does not GET the bytes during search.

### Contract probe, 2026-11-16

Identifying User-Agent `mcp-server-duckduckgo-research (+https://github.com/maccavelli/mcp-server-duckduckgo)`, `Accept: application/json`, at least 1.1s between calls. Every `/search.json` response was HTTP 200, `content-type: application/json` (no charset), `server: nginx/1.30.6`, `x-served-by: FastAPI`. No rate-limit header was present (`x-ratelimit-*`, `retry-after` absent). Top-level keys on every search body: `docs`, `documentation_url`, `numFound`, `numFoundExact`, `num_found`, `offset`, `q`, `start`. `numFound` and `num_found` were equal. `offset` was JSON `null` when the request omitted it.

| Request | Bytes | Time | What it showed |
| --- | --- | --- | --- |
| `q=python programming`, `limit=5`, planned fields including `isbn` and `publisher` | 1,863 | 688ms (Solr 226ms) | 5 docs, keys `/works/OL…W`. `cover_i`, `edition_count`, `first_publish_year`, `number_of_pages_median` were JSON integers. `author_name`, `isbn`, `publisher`, `language` were arrays. `subtitle` was a string when present and omitted otherwise. |
| `q=the lord of the rings`, `limit=5`, same fields | 38,206 | 397ms (Solr 17ms) | `/works/OL27448W`: 473 ISBNs, 113 publishers, 18 languages, `language[0]=fre`, `publisher[0]=Hör Verlag, München`, `first_publish_year=1954`, `number_of_pages_median=1193`, `edition_count=253`, `cover_i=14625765`. |
| Same query, fields without `isbn` and `publisher` | 1,803 | 350ms | Same five works. |
| `limit=0`, `fields=key,title` | 180 | 777ms | HTTP 200, `numFound=2557`, `docs=[]`. |
| `limit=1000`, `fields=key,title` | 76,732 | 1,444ms (Solr 1,014ms) | HTTP 200, 1,000 docs, all keys `/works/OL…W`. |
| `q=` / `q=zzzzqqqnonexistentbookxyz` | 156 / 181 | 330ms / 343ms | HTTP 200, `numFound=0`, `docs=[]`. |
| `fields=not_a_real_field_xyz` | 170 | 386ms | HTTP 200, one doc with no properties. Unknown fields are dropped, not a 422. |
| `http://openlibrary.org/search.json` | 221 | 260ms | HTTP 200 JSON. The final URL stayed `http://`. This client still uses HTTPS. |
| `GET /works/OL30906747W` | HTML | 1,101ms | 200, `text/html; charset=utf-8`, final URL `/works/OL30906747W/Core_Python_Programming`. |
| `GET /OL30906747W` | HTML | 441ms | 404. |
| `GET https://covers.openlibrary.org/b/id/13192647-M.jpg` (`cover_i` of the first python hit) | 20,820 | 1,382ms | 200 `image/jpeg`. Final URL was archive.org `view_archive.php` for `0013192647-M.jpg`. |

A scratch `encoding/json` program on `go1.27.4` unmarshalled `"offset": null` into an `int` as 0 with a nil error, and failed the parent value on `1193.5` into an `int` (`json: cannot unmarshal number 1193.5 into Go value of type int`). `number_of_pages_median` is decoded as `float64`. The other numeric fields matched JSON integers in this probe and stay `int`.

`cmd.Version` is `main.RawVersion`, set by `-X main.RawVersion` in the Makefile. The engine package cannot import `main`. The User-Agent therefore carries the module name and repository URL, not that ldflag.

### Implications for this MCP client

* One `GET https://openlibrary.org/search.json` per `search_books` call, with `q`, `limit` clamped to `[1, 50]` (values below 1 become 5), and `fields=key,title,subtitle,author_name,first_publish_year,number_of_pages_median,cover_i,edition_count`.
* `Accept: application/json`. The `.json` suffix already selects JSON; `Accept: */*` also returned `application/json` in the probe.
* User-Agent `mcp-server-duckduckgo (+https://github.com/maccavelli/mcp-server-duckduckgo)`. If `OPENLIBRARY_CONTACT` is set, append it inside that parenthesis. Do not invent an email. Build the request with `http.NewRequestWithContext`. `newRequest` calls `Store.GetUA`, which persists a browser User-Agent for the hostname.
* A blank query returns an error and does not use the rate-limit slot.
* Work URL is `https://openlibrary.org` + a `/works/…` key. A bare `OL…W` is prefixed with `/works/`. `https://openlibrary.org/` + a bare id 404s.
* `Description` and `Info` carry the same metadata line (authors, year, median pages, edition count) because `SearchResponse.ToMarkdown` quotes `Description` and ignores `Author`, `Date`, and `Info`.
* `Publisher` stays empty. There is no single publisher on a work.
* Source string `Open Library`.
* Do not call `/query.json`, `/api/things`, `/api/search`, `/api/books`, `/search/editions.json`, Infogami work/edition GETs, or the Covers host except as a constructed thumbnail URL.

Hub "Services using Open Library APIs" already lists `mcp-open-library` as an MCP consumer of this same catalog.
