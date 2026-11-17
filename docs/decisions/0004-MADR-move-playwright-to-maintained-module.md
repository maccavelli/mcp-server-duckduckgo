---
status: accepted
date: 2026-11-17
---
# Move playwright-go onto the maintained module path

## Context and Problem Statement

`internal/engine/playwright.go` imports `github.com/playwright-community/playwright-go` at v0.6000.0. That tag is the newest release whose `go.mod` still says `module github.com/playwright-community/playwright-go`.

On 2026-11-17 the module proxy lists `v0.6401.0` as the newest `playwright-go` tag. Its `go.mod` says `module github.com/mxschmitt/playwright-go` and `go 1.22`. The proxy records that tag at 2026-11-04T23:59:11Z, commit `f7cabe5851d98a3261d09336845c80756ef72142`. `https://github.com/playwright-community/playwright-go` answers 301 to `https://github.com/mxschmitt/playwright-go`.

`go get github.com/playwright-community/playwright-go@v0.6401.0` fails because the tag declares the mxschmitt path. The same mismatch is on v0.6100.0, v0.6201.0, and v0.6201.1. [0003-MADR-update-dependency-pins.md](0003-MADR-update-dependency-pins.md) therefore left the community pin at v0.6000.0. The owner then asked to move playwright to the maintained project and approved that move.

v0.6000.0 installs driver 1.60.0 from `https://playwright.azureedge.net` and `https://playwright-akamai.azureedge.net`. On 2026-11-17 both of these URLs returned HTTP 404:

* `https://playwright.azureedge.net/builds/driver/playwright-1.60.0-mac-arm64.zip`
* `https://playwright-akamai.azureedge.net/builds/driver/playwright-1.60.0-mac-arm64.zip`

`https://registry.npmjs.org/playwright/1.64.1` returned HTTP 200 the same day. v0.6401.0 installs its driver from the npm registry. Its `go.mod` does not require `go.mongodb.org/mongo-driver`. This module's current indirect requirement on `go.mongodb.org/mongo-driver v1.17.10` arrived with community v0.6000.0.

The call sites in `internal/engine/playwright.go` use `Install`, `Run`, `Stop`, `Chromium.Launch`, `BrowserTypeLaunchOptions.Headless`, `BrowserNewContextOptions` (`UserAgent`, `Viewport`, `Locale`, `TimezoneId`, `ColorScheme`), `ColorSchemeDark`, `AddInitScript`, `Script.Content`, `String`, `SetDefaultTimeout`, `Goto`, `WaitUntilStateDomcontentloaded`, `Content`, `Cookies`, and `Close`. Those symbols are present on v0.6401.0 with the same call shapes this file already uses, including `Content()` with no arguments.

Which module path does this server require?

## Decision Drivers

* The owner approved the move to the maintained module on 2026-11-17.
* A tag at v0.6100.0 or later resolves only as `github.com/mxschmitt/playwright-go`.
* The pinned 1.60.0 driver zip is absent from the CDN this tag still downloads.
* v0.6401.0 is the newest tag on the proxy list checked 2026-11-17.
* `go 1.27.4` stays. `github.com/ysmood/fetchup` stays at v0.3.0, because `go-rod` v0.116.2 still calls `fetchup.New(dir, urls...)`.
* Stealth injection, cookie field copies, and the five-second wait stay as they are.

## Considered Options

* Require `github.com/mxschmitt/playwright-go` at v0.6401.0 and change the one import
* Stay on `github.com/playwright-community/playwright-go` v0.6000.0
* Remove the Playwright fetch fallback

## Decision Outcome

Chosen option: "Require `github.com/mxschmitt/playwright-go` at v0.6401.0 and change the one import", because that is the maintained module, it is the newest tag, and the community pin cannot fetch its driver zip.

### Consequences

* Good, because `go get` can resolve v0.6401.0 on the path the tag declares.
* Good, because `Install()` fetches driver 1.64.1 from the npm registry, which answered HTTP 200.
* Good, because the symbols `internal/engine/playwright.go` calls still match.
* Neutral, because the first run after the bump downloads the 1.64.1 driver and browsers into the user cache. A previously cached 1.60 driver does not satisfy this tag.
* Neutral, because `go mod tidy` drops an indirect module that only the community pin required. `go.mongodb.org/mongo-driver` stayed: `golang-set/v2` v2.9.0 imports it.
* Bad, because a fresh machine needs network access to the npm registry and `nodejs.org` before the fallback can launch a browser.

### Confirmation

`go.mod` requires `github.com/mxschmitt/playwright-go v0.6401.0` and does not require `github.com/playwright-community/playwright-go`. The import in `internal/engine/playwright.go` is `github.com/mxschmitt/playwright-go`. `go` is still `1.27.4`. There is no `toolchain` line. `github.com/ysmood/fetchup` is still v0.3.0. `go build ./...` exits 0.

## Pros and Cons of the Options

### Require `github.com/mxschmitt/playwright-go` at v0.6401.0 and change the one import

* Good, because one import and one require line follow the module the repository now publishes.
* Good, because the driver install host that answered HTTP 200 is the one this tag uses.
* Bad, because the first `Install()` downloads a new driver and browser set.

### Stay on `github.com/playwright-community/playwright-go` v0.6000.0

* Good, because the tree already builds against that pin.
* Bad, because every newer tag rejects that import path.
* Bad, because the 1.60.0 mac-arm64 driver zip returned HTTP 404 on both CDN hosts checked 2026-11-17.

### Remove the Playwright fetch fallback

* Good, because the server would no longer download a browser driver.
* Bad, because CAPTCHA and HTTP 403 fetches would lose the fallback in `internal/engine/fetch.go`.

## More Information

This decision supersedes the playwright pin recorded in [0003-MADR-update-dependency-pins.md](0003-MADR-update-dependency-pins.md). The rest of that pin update stands, including fetchup v0.3.0.

### Amendment 2026-11-17 — mongo-driver stays

v0.6401.0's `go.mod` does not require `go.mongodb.org/mongo-driver`. `go mod tidy` still kept v1.17.10. `go mod why -m` traces `internal/engine` → `github.com/mxschmitt/playwright-go` → `github.com/deckarep/golang-set/v2` → `go.mongodb.org/mongo-driver/bson`. `golang-set` v2.9.0's `set.go` imports `go.mongodb.org/mongo-driver/bson/bsontype`, and that module's `go.mod` requires mongo-driver v1.17.4. MVS keeps v1.17.10. `github.com/go-jose/go-jose/v3` left the indirect block.

The Go floor stays the decision in [0002-MADR-pin-go-1.27.4.md](0002-MADR-pin-go-1.27.4.md).

Implementation plan: [0004-PLAN-move-playwright-to-maintained-module.md](0004-PLAN-move-playwright-to-maintained-module.md).
