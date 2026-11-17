# MADR and PLAN: load the skill, and gate mutating work

**Normative text lives in `AGENTS.md`**, section "MADR and PLAN before mutating
work". This file carries only the skill name and the gate. Do not restate the
workflow here.

## The skill

**Whenever the user asks for an MADR and a plan, load the
`madr-and-plan-writing` skill first** and follow it for authoring, naming
(`NNNN-MADR-*` / `NNNN-PLAN-*`) and review, for a fresh pair and for amending
an existing one.

The name is exact. A mistyped one returns `Unknown skill` and fails quietly.
The filesystem outranks this paragraph: read the `name:` line of
`~/.agents/skills/madr-and-plan-writing/SKILL.md` (or the repository's vendored
`.agents/skills/` copy) with your file-reading tool.

## The gate

**Mutating work needs an approved `docs/decisions/NNNN-MADR-*` /
`docs/decisions/NNNN-PLAN-*` pair first**; read-only investigation does not,
and neither does writing a `docs/reports/NNNN-REPORT-*`, which decides nothing.
What counts as mutating, the approval order, the bootstrap exception and the
pre-add checks are in `AGENTS.md`. Read them there rather than trusting a
summary.
