# MagicDuck MCP server

This repository is the source of truth for `mcp-server-duckduckgo`.

`mcp-server-duckduckgo` is a stdio MCP server for web, news, image, and video search,
Open Library book lookup, and page fetch. The current release is `v1.3.0`.
Published platforms are `linux/amd64`, `darwin/arm64`, and `windows/amd64`.

Decision records, the architecture note, and the guides are indexed in [docs/README.md](docs/README.md).

## Quick start

1. Install the latest release with one command.

   macOS or Linux:

   ```bash
   curl -fsSL https://github.com/maccavelli/mcp-server-duckduckgo/releases/latest/download/install.sh | sh
   ```

   Windows PowerShell:

   ```powershell
   irm https://github.com/maccavelli/mcp-server-duckduckgo/releases/latest/download/install.ps1 | iex
   ```

2. Start the server with no arguments. It speaks MCP over stdio.

   ```text
   mcp-server-duckduckgo
   ```

3. Connect a client. MagicTools launches this binary in production.
   A standalone client uses the command above.

Pinned tags and install options are in [Installing](docs/guides/installing.md).
Client setup is in [Connecting a client](docs/guides/connecting-a-client.md).

## Install

```bash
curl -fsSL https://github.com/maccavelli/mcp-server-duckduckgo/releases/latest/download/install.sh | sh
```

```powershell
irm https://github.com/maccavelli/mcp-server-duckduckgo/releases/latest/download/install.ps1 | iex
```

The installer checks `SHA256SUMS`, installs the binary to `~/.local/bin` or `%LOCALAPPDATA%\Programs\mcp-server-duckduckgo`, and runs `identity`. Options, pinned tags, and uninstall are in [docs/guides/installing.md](docs/guides/installing.md).

## Overview

- Searches the web, news, images, and videos.
- Looks up books with `search_books` against the Open Library Search API.
- Fetches a page with `fetch_url`.
- Speaks MCP over stdio. Starting the binary with no arguments runs the server.
- Serves in-memory diagnostic logs with `get_internal_logs`.

## Table of contents

- [Quick start](#quick-start)
- [Install](#install)
- [Overview](#overview)
- [Table of contents](#table-of-contents)
- [I want to…](#i-want-to)
- [Update](#update)

## I want to…

| I want to… | Start here |
| :--- | :--- |
| install the binary | [Installing](docs/guides/installing.md) |
| update an installed binary | [Updating](docs/guides/updating.md) |
| connect a client | [Connecting a client](docs/guides/connecting-a-client.md) |
| call a tool | [MCP tools](docs/guides/mcp-tools.md) |
| set the cache and environment | [Configuration](docs/guides/configuration.md) |
| build and test | [Developing](docs/guides/developing.md) |
| cut a release | [Releasing](docs/guides/releasing.md) |
| find a decision | [Documentation index](docs/README.md) |

## Update

```bash
mcp-server-duckduckgo update            # confirm, then install the latest release
mcp-server-duckduckgo update --check    # report only; exit 10 if an update is available
mcp-server-duckduckgo update --version v1.2.0   # install that exact release
```

Flags, exit codes, and local builds are in [docs/guides/updating.md](docs/guides/updating.md).
