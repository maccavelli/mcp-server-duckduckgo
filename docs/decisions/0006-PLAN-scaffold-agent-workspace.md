---
status: implemented
date: 2026-11-17
associated-madr: "0006-MADR-scaffold-agent-workspace.md"
---
# Implement the fleet agent workspace

Associated MADR: [0006-MADR-scaffold-agent-workspace.md](0006-MADR-scaffold-agent-workspace.md)

## Goal

A clone of this repository carries `CLAUDE.md`, an `AGENTS.md` symlink to it, the five harness pointer sets, the record checker, LF attributes, and the fleet Markdown config. ~~Records 0001 through 0005 are left unchanged.~~ Amended 2026-11-17: those records are brought onto the strict profile first, and the legacy filter stays empty.

## Scope

In scope: the file set in `workspace-scaffolding` `references/file-set.md`, the `.gitignore` merge, the `Makefile` splice, `docs/architecture.md`, index rows, and a link from the product `README.md` to `docs/README.md`.

Out of scope: Go source, tests, `go.mod`, CI workflows, product README prose other than that link, the `Makefile` Go recipes, and the `Makefile` comment that cites mcplib's build-kind record. No `.editorconfig`, `SECURITY.md`, `CODEOWNERS`, `CONTRIBUTING.md`, `CHANGELOG.md`, Dependabot, or permission allowlist. No empty `docs/guides/`.

## Grounding Reference

Templates are the copies under the installed `workspace-scaffolding` skill, directory `templates/`. Apply order is that skill's `references/apply.md`. This repository today:

* Missing: `CLAUDE.md`, `AGENTS.md`, `.gitattributes`, `opencode.json`, `scripts/check_records.py`, `docs/architecture.md`, and every pointer path in the file set.
* Present and kept: product `README.md`, `docs/README.md`, `docs/decisions/` 0001 through 0005, `Makefile` Go recipes, `.gitignore` product lines.
* Present and replaced by the template plus the legacy filter: `.markdownlint-cli2.jsonc`.
* Hooks path is the global hooks directory. Commits, when asked, use `git commit --no-edit`.

`{{SCOPE}}` text:

```text
MagicDuck is the MCP sub-server for DuckDuckGo web, news, image, and video search, plus Open Library book search and a page-fetch fallback. What exists today is docs/architecture.md. The decision index is docs/README.md.
```

`{{DEPENDENCIES}}` and `{{PREADD}}` are the template defaults. There is no `make preflight`. `{{WHAT_IT_IS}}` states the same product as fact: the `cmd/mcp-server-duckduckgo` binary, search and fetch in `internal/`, and one docs tree. `{{WHAT_IS_NOT_HERE}}` states that the operator cache and config live outside the repository, and that this public tree is a published export of a private source. No "we used to".

The legacy filter is exactly:

```text
"**/000[1-5]-{MADR,PLAN}-*.md"
```

Amended 2026-11-17: do not install that pattern. The template's legacy `filter` stays `[]`, because records 0001 through 0005 were edited to the strict profile.

## Prerequisites

* The skill templates are readable on this machine.
* `python3` is on `PATH`.
* Network is available for `npx --yes markdownlint-cli2@0.23.2` during verification.
* Working tree has no unrelated edits. This plan does not touch them if one appears; it stops.

## Phase 1 — Install the workspace

From the repository root, in the apply order:

1. Copy `templates/.gitattributes` to `.gitattributes` when the target is missing.
2. Append gitignore-fragment lines that are not already in `.gitignore`: `__pycache__/` and `*.py[cod]`.
3. Splice `templates/Makefile.snippet` into `Makefile`: the portable `SHELL` block and the `check-records` and `markdownlint` targets. Leave every existing Go recipe and `MOD_VERSION` in place.
4. Copy `templates/.markdownlint-cli2.jsonc` over `.markdownlint-cli2.jsonc`. ~~Set the legacy override `filter` to the pattern in Grounding Reference.~~ Leave the legacy `filter` empty. Change nothing else in that file. Amended 2026-11-17.
5. Copy `templates/scripts/check_records.py` to `scripts/check_records.py`.
6. Write `docs/architecture.md` from its template with the two slots filled. Do not create `docs/guides/`.
7. Add a link to `docs/README.md` in the product `README.md`. Do not replace the product text. Add the 0006 rows and an architecture row to the existing `docs/README.md` tables. Do not regenerate that index from a template.
8. Write `CLAUDE.md` from its template with the three slots filled. From the repo root, `ln -s CLAUDE.md AGENTS.md`.
9. Copy every pointer, both gitignores, and `templates/opencode.json` from the file set.

`cmp` each byte-for-byte file against its template. A difference other than the legacy filter is a bug. Stop.

Fix Markdown defects the human profile reports in files outside records, when the fix is a list marker, a table pipe, a second H1, or a blank line. Do not edit a fenced block inside a record. ~~Do not rewrite 0001 through 0005 to add the new PLAN headings.~~ Amended 2026-11-17: the owner directed that heading and status fix. It is the deviation below, and it lands before the file copies in this phase.

## Verification

```bash
python3 scripts/check_records.py --next
python3 scripts/check_records.py --check-all
npx --yes markdownlint-cli2@0.23.2
make check-records
make markdownlint
```

`--next` prints `0007`. `--check-all`, both Make targets, and markdownlint exit 0.

```bash
git ls-files --eol
git ls-files -s CLAUDE.md AGENTS.md
```

Read both outputs whole. Every tracked text file shows `attr/text=auto eol=lf`. `CLAUDE.md` is `100644`. `AGENTS.md` is `120000` and its target is exactly `CLAUDE.md`.

Plants run in a scratch copy, never in this tree. Show one `check_records.py --check-all` failure (`broken relative link:`) and one markdownlint record-profile failure (`MD043` on a PLAN that has no `## Completion Criteria`). Say what the failure looked like. Scan added files for the local account name, a mail host, and a home-directory prefix. Describe the scan. Do not write the identifier into a record.

## Rollback

Delete the added workspace files and restore `.gitignore`, `.markdownlint-cli2.jsonc`, `Makefile`, `README.md`, and `docs/README.md` from `HEAD`. ~~Records 0001 through 0005 are already on `HEAD` and stay.~~ Amended 2026-11-17: reverting the file-set copy does not revert the strict-profile edits to records 0001 through 0005 unless the owner asks to revert those edits too.

## Completion Criteria

* The file set is present and the byte-for-byte copies match their templates.
* ~~The legacy filter matches 0001 through 0005 and does not match this plan.~~ The legacy filter is empty. Records 0001 through 0006 pass the strict profile. Amended 2026-11-17.
* `AGENTS.md` is a relative symlink to `CLAUDE.md` in git mode `120000`.
* `check_records.py --check-all` and markdownlint exit 0.
* Both plants failed in a scratch copy, and the failure text was recorded here.
* No commit unless the owner asks in the same turn.

## Deviations

### 2026-11-17 — strict profile for records 0001 through 0005

The owner asked to lint and fix the old records so the docs gates pass when CI later runs `check_records.py --check-all` and `markdownlint-cli2@0.23.2` with the workspace template. That contradicts the first draft, which left 0001 through 0005 unchanged and listed them in the legacy filter.

Resolution: edit those records onto the strict profile and leave the legacy `filter` empty. No other phase-1 file is copied in this slice. The file-set install stays unapproved. The same scratch lint reported `MD022` and `MD032` on the product `README.md`. Blank lines were added under three headings and before two lists. No product sentence changed.

Later the same day the owner accepted the MADR, approved this plan, and said to proceed. Phase 1 then installed the file set. The legacy `filter` stayed empty.

A scratch copy of the template config, not this tree, showed the failures before the edits: five plans rejected for `status: complete`, five MADRs missing `## Consequences`, five plans missing `## Grounding Reference`, plus `MD038`, `MD040`, `MD010`, and `MD029` on the lines named in those plans' execution notes. The same scratch run showed `MD043` on a plan with no `## Completion Criteria`, and a clean pass on a plan that places Goal and Scope before `## Grounding Reference` and places `## Rollback` and `## Completion Criteria` after the execution record.

## Execution record

### 2026-11-17 — record-profile slice

Records 0001 through 0005 were edited as the deviation describes. Record 0006 already had the required headings. The product `README.md` gained the blank lines named in the deviation. At the end of that slice the plan was still `proposed` and nothing was committed. The owner then accepted the MADR, approved this plan, and said to proceed.

### 2026-11-17 — phase 1 installed the workspace

The owner asked to commit this phase on `main`. Push was not part of that ask.

Byte-for-byte copies matched their templates: `.gitattributes`, `.markdownlint-cli2.jsonc` (legacy `filter` left `[]`), `scripts/check_records.py`, both gitignores, `opencode.json`, and every pointer. Named edits: `.gitignore` gained `__pycache__/` and `*.py[cod]`; `Makefile` gained the portable `SHELL` block and the two gate targets; `CLAUDE.md` and `docs/architecture.md` are the templates with slots filled; `docs/README.md` gained an architecture row (the 0006 rows were already there); the product `README.md` links to `docs/README.md`. `AGENTS.md` is a relative symlink whose target is `CLAUDE.md`. `docs/guides/` was not created. The `Makefile` comment that cites mcplib's build-kind record was left in place.

`python3 scripts/check_records.py --next` printed:

```text
0007
```

`python3 scripts/check_records.py --check-all` printed nothing and exited 0. `make check-records` echoed `python3 scripts/check_records.py --check-all` and exited 0.

`npx --yes markdownlint-cli2@0.23.2` and `make markdownlint` each printed:

```text
markdownlint-cli2 v0.23.2 (markdownlint v0.41.1)
Linting: 27 files
Summary: 0 issues in 0 files
```

Scratch plants, outside this tree: `check_records.py --check-all` on a README whose only link is `[x](missing.md)` printed `broken relative link: README.md:3: missing.md` and `1 error(s)`, exit 1. markdownlint-cli2 0.23.2 on a plan with no `## Completion Criteria` printed `0098-PLAN-no-completion.md:22 error MD043/required-headings Required heading structure [Context: "*"]`, exit 1.

`git ls-files -s CLAUDE.md AGENTS.md`:

```text
120000 681311eb9cf453d0faddf3aacaec7357e97ba8e9 0 AGENTS.md
100644 75b6f73b5498bac5a449809f03a242384d027cf7 0 CLAUDE.md
```

`git ls-files --eol` was read in full: 98 paths. Every line contains `attr/text=auto eol=lf`. `AGENTS.md` is blank in the index and work-tree columns because it is a symlink; its attribute is still `text=auto eol=lf`.

Added and changed files were scanned for the local account name, a home-directory prefix, and a noreply mail host. The scan reported no hits. The identifier was not written into this record.
