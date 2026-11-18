# Installing

Task index: [docs/README.md](../README.md).

The `v1.2.0` release publishes `install.sh` and `install.ps1` with the
binaries and `SHA256SUMS`. Each installer installs the release it was
downloaded from. It does not ask the GitHub API which tag is latest.

## Latest release

macOS or Linux:

```bash
curl -fsSL https://github.com/maccavelli/mcp-server-duckduckgo/releases/latest/download/install.sh | sh
```

Windows PowerShell:

```powershell
irm https://github.com/maccavelli/mcp-server-duckduckgo/releases/latest/download/install.ps1 | iex
```

`irm … | iex` cannot pass parameters. Use the scriptblock form below when you need `-Version`, `-InstallDir`, or another switch.

## A specific tag

A tag URL installs that tag. Do not also pass `--version` to that downloaded script.

macOS or Linux:

```bash
curl -fsSL https://github.com/maccavelli/mcp-server-duckduckgo/releases/download/v1.2.0/install.sh | sh
```

Windows, when the script must take `-Version`:

```powershell
& ([scriptblock]::Create((irm https://github.com/maccavelli/mcp-server-duckduckgo/releases/latest/download/install.ps1))) -Version v1.2.0
```

## What the installer does

- Downloads `SHA256SUMS` and this platform's asset over HTTPS, then
  checks the file before anything is installed.
- Installs on Unix into `~/.local/bin`, without `sudo`. Root is refused
  unless `--allow-root` is set. On Windows it installs into
  `%LOCALAPPDATA%\Programs\mcp-server-duckduckgo` and updates the user PATH.
- Leaves the previous binary as `mcp-server-duckduckgo.prev` (`mcp-server-duckduckgo.exe.prev` on Windows).
- Runs `identity`. The installed binary must report the release stamp. If it does not, the previous binary is restored.
- Runs no install hooks. `selfupdate-release.json` sets `installer.name` and does not set hooks.

Supported platforms are `linux/amd64`, `darwin/arm64`, and `windows/amd64`.
`install.ps1` reads the machine architecture from the registry.
A Windows ARM64 machine is refused, because this release has no
`windows/arm64` asset. The error lists the platforms the release ships.

## Options

| install.sh | install.ps1 | Variable | What it does |
| :--- | :--- | :--- | :--- |
| `--version TAG` | `-Version TAG` | `MCP_SERVER_DUCKDUCKGO_VERSION` | Install another release. The tag may be written with or without a leading `v`. |
| `--dir DIR` | `-InstallDir DIR` | `MCP_SERVER_DUCKDUCKGO_INSTALL_DIR` | Install into that directory. A relative directory is made absolute first. |
| `--dry-run` | `-DryRun` | | Print what would be downloaded and where it would go. Change nothing. |
| `--uninstall` | `-Uninstall` | | Remove the binary and its `.prev` file from the install directory. Configuration files are left in place. |
| `--allow-root` | | | Allow `install.sh` to run as root. |
| `--verify-attestation` | `-VerifyAttestation` | | Also verify each download's attestation with a logged-in `gh`. |
| `-h`, `--help` | | | Print the options and exit 0. |

The variable prefix is `MCP_SERVER_DUCKDUCKGO`, the library default for the installer name `mcp-server-duckduckgo`. A switch takes no value: `-DryRun`, not `-DryRun:$true`.

Uninstall the same directory you installed into. Pass the same `--dir` or `-InstallDir`, or set `MCP_SERVER_DUCKDUCKGO_INSTALL_DIR`.

## Exit codes

| Code | Meaning |
| :--- | :--- |
| 0 | Installed, removed, or a dry run. |
| 1 | Usage or environment error: an unknown option, an unsupported platform, root without `--allow-root`, a missing tool, or a refused version. |
| 2 | Verification failure: a download, `SHA256SUMS`, a checksum, an attestation, or the identity check. Nothing was changed, or the previous binary was restored. |
| 3 | Removed only in part, because a program could not be removed. |

Through `iex` or the scriptblock form, `install.ps1` does not exit the PowerShell session. A failure is the error `install failed (exit N)`.
