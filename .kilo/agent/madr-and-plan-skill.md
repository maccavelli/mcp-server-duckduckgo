# MADR and PLAN: load the skill, and gate mutating work

**Normative text lives in `AGENTS.md`**, section "MADR and PLAN before mutating
work". This file carries only the two things that must not be missed if that
file is not loaded. Do not restate the workflow here.

## The skill

Whenever the user asks for an MADR and a plan, load the **`madr-and-plan-writing`**
skill first and follow it for authoring, naming (`NNNN-MADR-*` / `NNNN-PLAN-*`)
and review, for a fresh pair and for amending an existing one.

The name is exact. A mistyped one returns `Unknown skill` and fails quietly.
The filesystem outranks this paragraph: read the `name:` line of
`~/.agents/skills/madr-and-plan-writing/SKILL.md` (or the repository's vendored
`.agents/skills/` copy) with your file-reading tool.

## The gate

**Read-only investigation needs no pair.** Reading, searching, `git log` /
`show` / `diff`, and existing tests or diagnostics that do not write the tree.

**Reports need no pair.** A `docs/reports/NNNN-REPORT-*` decides nothing;
writing one, with its index row, is not gated. Acting on what it recommends is.

**Mutating work does.** Before the first write, name the
`docs/decisions/NNNN-MADR-*` / `docs/decisions/NNNN-PLAN-*` pair being
executed, or stop and write one. See `AGENTS.md` for what counts as mutating,
the approval order, the bootstrap exception, and the pre-add checks.
