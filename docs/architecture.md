# Architecture

How this repository is put together, as it is now. This file carries no
history and no rationale — when you want to know *why* something is this
way, the decision records under [decisions/](decisions/) hold the argument,
and [README.md](README.md) points at the ones people ask for most.

## What it is

The `cmd/mcp-server-duckduckgo` binary is the MCP sub-server. Search and page
fetch live under `internal/`. Book search calls the Open Library Search API.
This repository has one documentation tree, `docs/`.

## Tree

```text
README.md                                        repository entry; links to docs/README.md
CLAUDE.md                                        the one project instruction file, read by every agent
AGENTS.md                                        symlink to CLAUDE.md, for harnesses that read AGENTS.md
.claude/rules/madr-and-plan-skill.md             Claude pointer: load madr-and-plan-writing, and the mutate gate
.claude/.gitignore                               keeps Claude's settings.local.json out of git
.grok/rules/madr-plan-before-mutating-work.md    Grok pointer
.codex/rules/madr-plan-before-mutating-work.md   Codex pointer
.kilo/agent/madr-and-plan-skill.md               Kilo pointer
.kilo/.gitignore                                 keeps Kilo's node_modules, lockfiles and agent-manager.json out of git
.opencode/rules.md                               OpenCode pointer
opencode.json                                    tells OpenCode to read .opencode/rules.md
.markdownlint-cli2.jsonc                         fleet Markdown lint: human style, plus the record profile
.gitattributes                                   LF line endings on every platform; common binary types marked binary
.gitignore                                       Python caches and scratch output, merged with this repository's own entries
scripts/check_records.py                         record and documentation checks; see Gates
docs/
  README.md                                      ToC and the "I want to…" matrix
  architecture.md                                this file
  decisions/                                     MADR/PLAN pairs
  reports/                                       numbered observations (created with the first report)
  guides/                                        unnumbered how-to documents (created with the first guide)
```

Where the repository has a `Makefile`, it also carries the `check-records` and `markdownlint` targets
that run the gates below.

## Gates

Run before every commit. A file that fails is not committed.

| Command | Checks |
| :--- | :--- |
| `python3 scripts/check_records.py --next` | Prints the next record number: one more than the highest anywhere in the repository, so a gap stays a gap. |
| `python3 scripts/check_records.py --check-all` | Record filenames and placement; MADR, PLAN and GATES pairing; MADR and PLAN frontmatter `status` and `date`, and each status against its kind's vocabulary; every relative link and `#anchor` in every record. Links in the root `README.md`, `docs/README.md`, `docs/architecture.md` and `docs/guides/`. No unfilled `{{…}}` slot in those files or in `CLAUDE.md`. |
| `npx --yes markdownlint-cli2@0.23.2` | Every Markdown file. Documents get the human style rules. MADRs and PLANs get the record profile: required sections in order, sequential steps, tagged fences, well-formed tables. |

`make check-records` and `make markdownlint` run the last two where the `Makefile` targets exist.

Git-ignored paths, and any directory whose name starts with `_` (a tree kept as a fallback after it was deprecated or moved), are outside every gate, including `--next`.

No gate checks links in `CLAUDE.md`, or in any `README.md` or `architecture.md` below the root tree.

## What is not here

- CI. The gates run by hand before each commit.
- Secrets of any kind.
- The host layer, which each clone gets from the operator's machine, not from this repository:
  - the global `prepare-commit-msg` hook that writes commit messages;
  - the pre-push disclosure guard, where installed;
  - the user skills the pointers load (`madr-and-plan-writing`, `documentation-writing`);
  - each agent's permission settings.
- Operator cache and config live outside this repository.
- This public tree is a published export of a private source.
