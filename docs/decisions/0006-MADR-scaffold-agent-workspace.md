---
status: accepted
date: 2026-11-17
---
# Install the fleet agent workspace

## Context and Problem Statement

This repository is an existing product. It has a Go module, a `Makefile`, a product `README.md`, `docs/decisions/` records 0001 through 0005, and a hand-written `docs/README.md`. It has no `CLAUDE.md`, no `AGENTS.md`, no harness pointers, no `docs/architecture.md`, no `scripts/check_records.py`, and no `.gitattributes`.

`.markdownlint-cli2.jsonc` is an older fleet README config. It excludes every file whose name contains `MADR` or `PLAN`, so the record profile never runs. The current workspace template lints those files with a record profile and keeps one per-repository delta: a legacy filter for records that existed before adoption.

`git rev-parse --path-format=absolute --git-path hooks` is the global hooks directory, so commit messages stay on `git commit --no-edit`.

There is one documentation tree. Search, fetch, and the book catalog are the product, not adjacent stacks. `docs/guides/` does not exist and is not created empty.

How does an agent clone get the same workspace gates as the rest of the fleet?

## Decision Drivers

* The owner asked to load `workspace-scaffolding` and bring this repository into compliance.
* Records 0001 through 0005 stay as written. Their PLAN headings predate the record profile.
* The product README, the Go recipes, and CI stay product-owned.
* Copied template files stay byte-identical to the skill templates. Named deltas are the only edits.
* One tree. No empty `docs/guides/`.

## Considered Options

* Install the workspace file set, with records 0001 through 0005 on the legacy lint filter
* Replace `docs/README.md` and the existing Markdown config with the templates and rewrite 0001 through 0005 to the new PLAN headings
* Leave the repository without `CLAUDE.md` and the record gate

## Decision Outcome

Chosen option: "Install the workspace file set, with records 0001 through 0005 on the legacy lint filter", because the product and its decisions are already here, and the missing piece is the agent layer.

Amendment 2026-11-17: the owner directed records 0001 through 0005 onto the strict record profile so a later docs gate can pass. The legacy filter stays empty. The chosen-option sentence above is the proposal as first written.

## Consequences

* Good, because every harness gets the same mutate-only-with-a-pair rule, and `AGENTS.md` cannot drift from `CLAUDE.md`.
* ~~Good, because 0001 through 0005 keep their headings. The legacy filter is the template's named delta for a repo that already has records.~~ Superseded 2026-11-17. Those records are edited onto the strict profile, and the legacy filter stays empty.
* Good, because records 0001 through 0006 share one strict profile, so the template filter can stay empty.
* Neutral, because `docs/README.md` stays the hand-written index. The scaffold adds rows. It does not replace the file.
* Bad, because a later record must satisfy the strict record profile, including a PLAN's `## Grounding Reference`, `## Prerequisites`, a stage heading, `## Rollback`, and `## Completion Criteria`.

## Confirmation

`python3 scripts/check_records.py --check-all` exits 0. `npx --yes markdownlint-cli2@0.23.2` exits 0. `git ls-files -s` shows `CLAUDE.md` as `100644` and `AGENTS.md` as `120000` pointing at `CLAUDE.md`. Every copied file in the file set is `cmp`-identical to its template. ~~The legacy filter matches records 0001 through 0005 and does not match 0006.~~ Amended 2026-11-17: the legacy filter is empty, and records 0001 through 0006 pass the strict profile.

## Pros and Cons of the Options

### Install the workspace file set, with records 0001 through 0005 on the legacy lint filter

* Good, because the apply order is the skill's, and the only config edit is the filter the template calls the per-repository delta.
* Good, because product source, tests, and CI are outside this change.
* Neutral, because the first real guide is what creates `docs/guides/`.

### Replace `docs/README.md` and the existing Markdown config with the templates and rewrite 0001 through 0005 to the new PLAN headings

* Good, because one heading style would cover every PLAN.
* Bad, because it rewrites accepted records and replaces an index that already names them.

### Leave the repository without `CLAUDE.md` and the record gate

* Good, because nothing in the product binary changes.
* Bad, because a clone still has no shared instruction file, no record checker, and a Markdown config that skips every decision.

## More Information

Skill `workspace-scaffolding`. Writer skills `documentation-writing` and `madr-and-plan-writing`, verified by the `name:` line on disk.

The `Makefile` comment that says "MADR 0005" names mcplib's build-kind record. This repository's 0005 is the golangci-lint pin. That comment stays.

Implementation plan: [0006-PLAN-scaffold-agent-workspace.md](0006-PLAN-scaffold-agent-workspace.md).

### Amendment 2026-11-17 — records 0001 through 0005 meet the strict profile

The owner asked to lint and fix the existing records so `python3 scripts/check_records.py --check-all` and `npx --yes markdownlint-cli2@0.23.2` pass once this repository's CI uses the workspace template. That replaces the legacy-filter exemption.

What changed in those records, and only that:

* Each finished plan's frontmatter `status` is `implemented`. The checker rejects `complete`.
* Each MADR's Consequences heading is an H2, in the required order.
* Each plan gained `## Grounding Reference`, `## Prerequisites`, `## Rollback`, and `## Completion Criteria`. Existing goal, scope, steps, verification, rollout, deviation, and execution prose stayed.
* Record-profile findings in those files were fixed: untagged fences, hard tabs, a code span with a trailing space, and an ordered list that a column-0 fence had split.

The product `README.md` needed blank lines under three headings and before two lists (`MD022`, `MD032`). Those lines were added. No product sentence changed.

The file-set copy in the plan is still unapproved. This amendment does not install `CLAUDE.md`, the pointers, or the Markdown config.
