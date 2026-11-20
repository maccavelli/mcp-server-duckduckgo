# CLAUDE.md

> **Authority.** `~/.claude/CLAUDE.md` governs every repository and is not
> relaxed or tightened by anything here. This file is the authoritative project
> instruction file for this repository, and `AGENTS.md` is a symlink to this file:
> one source, discovered under both names. Do not replace the symlink with a
> separate file.

Instructions for AI coding agents working in this repository. All agents read
this file, under either name. A harness pointer in `.claude/rules/` /
`.grok/rules/` / `.opencode/rules.md` / `.codex/rules/` / `.kilo/agent/` wins
only where it is more specific than this file.

MagicDuck is the MCP sub-server for DuckDuckGo web, news, image, and video
search, plus Open Library book search and a page-fetch fallback. What exists
today is `docs/architecture.md`. The decision index is `docs/README.md`.

## Dependencies

No dependency may be added without a MADR in this repository that names it.
Lockfiles change in the same commit as the first import, and are cleaned in
the commit that removes the last. The first stack record owns the exact
commands.

## MADR and PLAN before mutating work

**Whenever the user asks for an MADR and a plan, load the
`madr-and-plan-writing` skill first** and follow it for authoring, naming and
review. This applies both to writing a fresh pair and to amending an existing
one.

The name is exact — it is the `name:` field of the skill. A mistyped call
returns `Unknown skill`, and an agent that proceeds without the skill writes
something shaped like a MADR while missing the required headings and the
directory rule below. Verify against the filesystem rather than memory: read the `name:` line of
the skill's `SKILL.md` (under `~/.claude/skills/`, `~/.agents/skills/`, or
the repository's vendored `.agents/skills/`) with your file-reading tool.

**Read-only investigation is allowed with no pair.** Reading, searching,
`git log` / `git show` / `git diff`, and existing tests or diagnostics that
do not write the tree do not need a MADR.

**Reports need no pair.** A `docs/reports/NNNN-REPORT-*` records what was
observed and decides nothing, so writing, amending, staging or committing one,
with its row in `docs/README.md`, is not gated. Acting on what a report
recommends is.

**Mutating work is not.** Before the first write, name the
`docs/decisions/NNNN-MADR-*` / `docs/decisions/NNNN-PLAN-*` pair being
executed, or stop and write one.

Mutating means: creating, editing, or deleting files other than a report;
staging or committing (except a report, or the bootstrap exception below);
dependency or lockfile changes;
CI / config / hook changes; builds or installers that write the tree,
`$HOME`, or a live service; generating committed artifacts.

Order:

1. Investigate (read-only).
2. Write or amend the MADR (`status: proposed` unless the owner already
   decided). Present it. Do not implement.
3. Write or amend the PLAN. Present it.
4. Mutate **only after** the owner explicitly approves execution
   (`proceed`, `execute the plan`, `do phase N`). Stay inside that PLAN.
5. Anything discovered mid-execution that is out of scope waits: amend the
   pair, re-approve, then continue. Completing a phase is not permission to
   invent the next unwritten one.

Follow-up vs greenfield:

- **Same topic** (debug, leftover phase, bug found in that plan's live
  run): amend that number. Add a PLAN phase or an amendment in the MADR. Do
  not silently rewrite historical rationale.
- **Greenfield**: next unused `NNNN`, new MADR, new PLAN, same slug. No
  mutation until that PLAN is approved.

Bootstrap exception: authoring `docs/decisions/NNNN-MADR-*`,
`docs/decisions/NNNN-PLAN-*`, this file, and the per-agent pointers
(`.claude/rules/`, `.grok/rules/`, `.opencode/rules.md`, `.codex/rules/`,
`.kilo/agent/`) does not require a *prior* pair. Putting source, tests, CI,
or product config in that same commit is a violation.

`git push` and tags still need an explicit ask in the same turn.

When the user asks where a document belongs, load `documentation-writing`.

## Records

```text
docs/decisions/NNNN-MADR-short-slug.md
docs/decisions/NNNN-PLAN-short-slug.md
docs/reports/NNNN-REPORT-short-slug.md
docs/reports/NNNN-GATES-short-slug.md
```

- `NNNN` is a zero-padded 4-digit number, one sequence across
  `docs/decisions/` and `docs/reports/`. A MADR and its PLAN share the same
  number and the same slug.
- **Next number** comes from `python3 scripts/check_records.py --next`.
  Never reuse a number, never renumber an existing record, never leave a gap
  deliberately.
- `make check-records` validates naming, pairing, placement, and every
  relative link in the records and the unnumbered docs. If Make is absent,
  run `python3 scripts/check_records.py --check-all`.
- Cite records by full filename, never by number alone. Cite another
  repository's record by repository and filename; a relative link cannot reach
  it.
- `docs/README.md` indexes every record by hand. Update it in the same change.

## Pre-add checks

Before staging Go files, run:

```bash
make pre-add-check                 # every tracked Go file
make pre-add-check FILES="a.go b.go"
```

It runs `scripts/go-precheck.sh`: `gofmt` on the files; `golangci-lint run -c
.golangci.yml ./...` once for the host GOOS, the same command, config and GOOS
as `make lint` (`docs/decisions/0005-MADR-pin-golangci-lint-v2.15.0.md` pins
the version); `go vet` and `go test` on the packages the files belong to; and
`govulncheck ./...` through the same pinned invocation as the `vuln` target.
`GO_PRECHECK_SKIP_VULN=1` skips govulncheck offline; `GOLANGCI_LINT=<path>`
names the golangci-lint binary, as for `make lint`.

`golint` is not used: the fleet retired the per-file rule (dotfiles
`docs/decisions/0028-MADR-retire-per-file-golint-from-the-fleet-pre-add-rule.md`),
and this repository never ran it. This repository's `revive` ruleset is an
explicit list (`.golangci.yml` under `revive.rules`) that does not enable
golint's `exported`, `package-comments` or `var-naming` rules, so nothing
here depended on them. A file that fails is not committed.

The machine-wide agent gate runs the same script before every agent
`git commit` in this repository.

- Before every commit, run `make check-records` and `make markdownlint` when
  those targets exist; otherwise `python3 scripts/check_records.py --check-all`
  and `npx --yes markdownlint-cli2@0.23.2`. A file that fails is not committed.

## Identifiers

Nothing committed carries a hostname, account name, org-internal path, or a
real-machine absolute path. Use placeholders (`<user>`, `/home/<user>/...`).

- When recording an identifier scan in a record or a commit, describe what was
  scanned for ("the local account name", "the hostname domain"); never quote
  it.
- Where a global pre-push disclosure guard is installed
  (`~/.global-git-hooks/github-disclosure.py`), run it over the outgoing
  commits before asking the owner to push. Feed it the ref line it expects on
  stdin from a Python file, not a shell pipeline.

- Never bypass it with `--no-verify`.

## Commits

`git commit --no-edit`. Never pass `-m` / `--message` / `-F`. The global
`prepare-commit-msg` hook writes the message from the staged diff.
