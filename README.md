> **Mirror notice:** This repository is a one-way published export of a
> privately hosted project. History is squashed into sync snapshots, and pull
> requests cannot be merged here directly — open an issue instead. Changes
> land in the private source and are re-exported.

<!-- markdownlint-disable MD013 MD060 MD033 -->

# MagicDuck MCP Sub-Server

A high-performance Model Context Protocol (MCP) sub-server providing secure web search and media discovery via DuckDuckGo.

## Keeping it up to date

```bash
mcp-server-duckduckgo update            # confirm, then install the latest release
mcp-server-duckduckgo update --check    # report only; exit 10 if an update is available
mcp-server-duckduckgo update --version v1.1.0   # install that exact release
```

`--check` reports only (exit `0` up to date, `10` actionable, `1` error) and
contradicts `--yes`/`--force`. `--yes`/`-y` approves without prompting; a
non-interactive apply without it fails rather than hanging. `--force` replaces a
locally built binary or reinstalls the selected version, and never bypasses
version, asset, size, integrity or target checks. `--version vX.Y.Z` installs an
exact release; a lower tag is reported as an explicit rollback.

`update` creates no cache directory or `config.yaml` and starts no datastore,
browser, search engine, tool registry, transport or MCP server. Its output goes
to stderr so the JSON-RPC stdout stays protocol-clean. Set `GH_TOKEN` or
`GITHUB_TOKEN` to raise GitHub API rate limits; the token is sent only to the
GitHub API origin.

Supported platforms are `linux/amd64`, `darwin/arm64` and `windows/amd64`. A
binary you built yourself is a local build and `update` refuses to replace it
without `--force`. If you installed through a package manager, update through
that manager instead — self-update does not take ownership of package-manager
installs.

Releases are immutable `vMAJOR.MINOR.PATCH` tags publishing the exact
`mcp-server-duckduckgo-<goos>-<goarch>[.exe]` assets and one `SHA256SUMS`. A
rebuilt fix gets a new patch tag rather than replacing a published asset
(mcplib MADR 0005).

## 🏗️ Core Pillars: Why, What, and How

### 1. Why the Server Exists
AI models suffer from training cutoff limits and lack real-time visibility into the public web, documentation updates, and breaking news. Commercial search APIs often require complex api-key signups, implement severe rate limits, and track queries. `mcp-server-duckduckgo` exists to give AI agents a zero-configuration, privacy-preserving gateway to retrieve live web context, documentation, and media assets using DuckDuckGo.

### 2. What the Server Does
The DuckDuckGo sub-server enables search and web scraping workflows:
- **Web & News Search**: Searches active websites and current news feeds.
- **Media Discovery**: Finds related images and videos.
- **Direct Web Scraping**: Fetches webpage content and dynamically converts it into clean, readable Markdown for prompt injection.
- **Search Suggestions**: Retrieves search query autocompletion suggestions.

### 3. How it Does What it Does
- **HTML Parsing & Extraction**: Queries DuckDuckGo HTML endpoints directly, parsing results concurrently with Go's `net/http` client.
- **Jit Web Reader**: The `read_url` tool fetches remote HTML pages, strips script/style tags, and structures the content as Markdown.
- **Telemetry Logging**: Writes process lifecycle logs to a high-fidelity ring buffer and routes standard error streams safely to the orchestrator.

---

## Quick Start

### Step 1: Place the Binary

Download the `mcp-server-duckduckgo` binary for your platform and place it in a directory on your system `PATH`.

#### Linux

```bash
# Move the binary to your local bin directory
mv mcp-server-duckduckgo ~/.local/bin/mcp-server-duckduckgo
chmod +x ~/.local/bin/mcp-server-duckduckgo
```

#### macOS

```bash
# Move the binary to your local bin directory
mv mcp-server-duckduckgo /usr/local/bin/mcp-server-duckduckgo
chmod +x /usr/local/bin/mcp-server-duckduckgo
```

#### Windows (PowerShell)

```powershell
# Create a directory for the binary if it doesn't exist
New-Item -ItemType Directory -Force -Path "$env:LOCALAPPDATA\Programs\duckduckgo"

# Move the binary
Move-Item mcp-server-duckduckgo.exe "$env:LOCALAPPDATA\Programs\duckduckgo\mcp-server-duckduckgo.exe"

# Add to your PATH (current user, persistent)
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
[Environment]::SetEnvironmentVariable("Path", "$currentPath;$env:LOCALAPPDATA\Programs\duckduckgo", "User")
```

---

### Step 2: Initialize Configuration

`mcp-server-duckduckgo` is a stateless sub-server. It does not require local configuration files, API tokens, or initialization steps like `init` or `configure`.

---

### Step 3: Configure Your IDE

> **⚠️ IMPORTANT ORCHESTRATOR MESSAGING**
>
> While the standalone IDE configurations below are provided for testing and debugging, `mcp-server-duckduckgo` is designed to be run as a downstream node behind the **`magictools` orchestrator** in production environments.
>
> When running in production, you should **only** configure `magictools` in your IDE, which will automatically proxy requests to `duckduckgo` as needed.

If you are testing the server standalone, configure your IDE to launch the binary directly (no `serve` argument is needed):

#### Antigravity (Google DeepMind)

| OS | Configuration File Path |
|---|---|
| Linux | `~/.gemini/antigravity/mcp_config.json` |
| macOS | `~/.gemini/antigravity/mcp_config.json` |
| Windows | `%USERPROFILE%\.gemini\antigravity\mcp_config.json` |

**Linux / macOS:**

```json
{
  "mcpServers": {
    "duckduckgo": {
      "command": "/home/youruser/.local/bin/mcp-server-duckduckgo",
      "env": {
        "HOME": "/home/youruser"
      }
    }
  }
}
```

**Windows:**

```json
{
  "mcpServers": {
    "duckduckgo": {
      "command": "C:\\Users\\YourUser\\AppData\\Local\\Programs\\duckduckgo\\mcp-server-duckduckgo.exe",
      "env": {
        "USERPROFILE": "C:\\Users\\YourUser"
      }
    }
  }
}
```

#### Visual Studio Code (GitHub Copilot / Native MCP)

| OS | User-Level Configuration File Path |
|---|---|
| Linux | `~/.config/Code/User/mcp.json` |
| macOS | `~/Library/Application Support/Code/User/mcp.json` |
| Windows | `%APPDATA%\Code\User\mcp.json` |

**Linux:**

```json
{
  "mcpServers": {
    "duckduckgo": {
      "command": "/home/youruser/.local/bin/mcp-server-duckduckgo",
      "env": {
        "HOME": "/home/youruser"
      }
    }
  }
}
```

**macOS:**

```json
{
  "mcpServers": {
    "duckduckgo": {
      "command": "/usr/local/bin/mcp-server-duckduckgo",
      "env": {
        "HOME": "/Users/youruser"
      }
    }
  }
}
```

**Windows:**

```json
{
  "mcpServers": {
    "duckduckgo": {
      "command": "C:\\Users\\YourUser\\AppData\\Local\\Programs\\duckduckgo\\mcp-server-duckduckgo.exe",
      "env": {
        "USERPROFILE": "C:\\Users\\YourUser"
      }
    }
  }
}
```

#### VSCode — Cline Extension

| OS | Configuration File Path |
|---|---|
| Linux | `~/.cline/data/settings/cline_mcp_settings.json` |
| macOS | `~/Library/Application Support/Code/User/globalStorage/saoudrizwan.claude-dev/settings/cline_mcp_settings.json` |
| Windows | `%APPDATA%\Code\User\globalStorage\saoudrizwan.claude-dev\settings\cline_mcp_settings.json` |

**Linux:**

```json
{
  "mcpServers": {
    "duckduckgo": {
      "command": "/home/youruser/.local/bin/mcp-server-duckduckgo",
      "env": {
        "HOME": "/home/youruser"
      }
    }
  }
}
```

**macOS:**

```json
{
  "mcpServers": {
    "duckduckgo": {
      "command": "/usr/local/bin/mcp-server-duckduckgo",
      "env": {
        "HOME": "/Users/youruser"
      }
    }
  }
}
```

**Windows:**

```json
{
  "mcpServers": {
    "duckduckgo": {
      "command": "C:\\Users\\YourUser\\AppData\\Local\\Programs\\duckduckgo\\mcp-server-duckduckgo.exe",
      "env": {
        "USERPROFILE": "C:\\Users\\YourUser"
      }
    }
  }
}
```

#### Claude Desktop

| OS | Configuration File Path |
|---|---|
| macOS | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| Windows | `%APPDATA%\Claude\claude_desktop_config.json` |

**macOS:**

```json
{
  "mcpServers": {
    "duckduckgo": {
      "command": "/usr/local/bin/mcp-server-duckduckgo",
      "env": {
        "HOME": "/Users/youruser"
      }
    }
  }
}
```

**Windows:**

```json
{
  "mcpServers": {
    "duckduckgo": {
      "command": "C:\\Users\\YourUser\\AppData\\Local\\Programs\\duckduckgo\\mcp-server-duckduckgo.exe",
      "env": {
        "USERPROFILE": "C:\\Users\\YourUser"
      }
    }
  }
}
```

#### Claude Code (CLI)

Claude Code uses a CLI command to register MCP servers.

**Linux:**

```bash
claude mcp add duckduckgo -s user -- /home/youruser/.local/bin/mcp-server-duckduckgo
```

**macOS:**

```bash
claude mcp add duckduckgo -s user -- /usr/local/bin/mcp-server-duckduckgo
```

**Windows (PowerShell):**

```powershell
claude mcp add duckduckgo -s user -- "C:\Users\YourUser\AppData\Local\Programs\duckduckgo\mcp-server-duckduckgo.exe"
```

#### Cursor

| OS | Global Configuration File Path |
|---|---|
| Linux | `~/.cursor/mcp.json` |
| macOS | `~/.cursor/mcp.json` |
| Windows | `%USERPROFILE%\.cursor\mcp.json` |

**Linux:**

```json
{
  "mcpServers": {
    "duckduckgo": {
      "command": "/home/youruser/.local/bin/mcp-server-duckduckgo",
      "env": {
        "HOME": "/home/youruser"
      }
    }
  }
}
```

**macOS:**

```json
{
  "mcpServers": {
    "duckduckgo": {
      "command": "/usr/local/bin/mcp-server-duckduckgo",
      "env": {
        "HOME": "/Users/youruser"
      }
    }
  }
}
```

**Windows:**

```json
{
  "mcpServers": {
    "duckduckgo": {
      "command": "C:\\Users\\YourUser\\AppData\\Local\\Programs\\duckduckgo\\mcp-server-duckduckgo.exe",
      "env": {
        "USERPROFILE": "C:\\Users\\YourUser"
      }
    }
  }
}
```

#### JetBrains IDEs (IntelliJ, GoLand, WebStorm, PyCharm)

JetBrains IDEs configure MCP servers via the AI Assistant settings or a local configuration file.

| OS | Configuration File Path |
|---|---|
| Linux | `~/.config/JetBrains/AI/mcp.json` (or via UI: Settings > Tools > AI Assistant > MCP Servers) |
| macOS | `~/Library/Application Support/JetBrains/AI/mcp.json` (or via UI: Settings > Tools > AI Assistant > MCP Servers) |
| Windows | `%APPDATA%\JetBrains\AI\mcp.json` (or via UI: Settings > Tools > AI Assistant > MCP Servers) |

**Linux:**

```json
{
  "mcpServers": {
    "duckduckgo": {
      "command": "/home/youruser/.local/bin/mcp-server-duckduckgo",
      "env": {
        "HOME": "/home/youruser"
      }
    }
  }
}
```

**macOS:**

```json
{
  "mcpServers": {
    "duckduckgo": {
      "command": "/usr/local/bin/mcp-server-duckduckgo",
      "env": {
        "HOME": "/Users/youruser"
      }
    }
  }
}
```

**Windows:**

```json
{
  "mcpServers": {
    "duckduckgo": {
      "command": "C:\\Users\\YourUser\\AppData\\Local\\Programs\\duckduckgo\\mcp-server-duckduckgo.exe",
      "env": {
        "USERPROFILE": "C:\\Users\\YourUser"
      }
    }
  }
}
```

---

## ⚙️ Configuration & Environment Variables

`mcp-server-duckduckgo` is stateless and does not require local configuration files. You can configure execution behavior using the following environment variables:

| Variable | Default | Description |
|---|---|---|
| `MCP_ENDPOINT_API_PORT` | `47688` | Local HTTP endpoint port for secondary diagnostic streams. |
| `MCP_LOG_LEVEL` | `INFO` | Verbosity of the server logs (`DEBUG`, `INFO`, `WARN`, `ERROR`). |

---

## 🛠️ MCP Tools Reference

Once the server is running, the following tools are exposed:

| Tool | Parameters | Description |
|---|---|---|
| `search_web` | `query` (string), `limit` (integer, optional) | **[DIRECTIVE: Web Search Leg]** Performs a standard search engine query and returns titles, snippets, and URLs. |
| `search_news` | `query` (string), `limit` (integer, optional) | **[DIRECTIVE: News Engine Leg]** Fetches current news headlines and links related to the query. |
| `search_images` | `query` (string), `limit` (integer, optional) | **[DIRECTIVE: Image Scanner]** Returns metadata and URLs for images matching the query. |
| `search_videos` | `query` (string), `limit` (integer, optional) | **[DIRECTIVE: Video Scanner]** Returns metadata and URLs for videos matching the query. |
| `search_suggest` | `query` (string) | **[DIRECTIVE: Autocomplete Prober]** Returns search suggestions/completions for the query string. |
| `read_url` | `url` (string) | **[DIRECTIVE: Webpage Ingestor]** Downloads a public URL and converts its HTML structure into clean Markdown. |
| `get_internal_logs` | None | **[SERVER: duckduckgo]** Fetches the tail of in-memory logs for diagnostic purposes. |

---

## 📋 Data Storage Locations

| Data | Linux | macOS | Windows |
|---|---|---|---|
| **Server Logs** | `stderr` (captured by IDE) | `stderr` (captured by IDE) | `stderr` (captured by IDE) |

---

*Built with ❤️. Part of the MagicTools Intelligence Suite.*
