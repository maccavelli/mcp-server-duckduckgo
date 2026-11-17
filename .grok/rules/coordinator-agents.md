# Coordinator agents before running a fan-out

Normative text lives in the `coordinator-agents` skill. This file carries only the
trigger and the one behavior that must not be missed if that skill is not loaded.
Do not restate the skill here.

## The skill

**Whenever you create a coordinator agent, act as one yourself, or collect and
verify results from several agents running in parallel, load the
`coordinator-agents` skill first** and follow its persona, role, and behavior
defaults.

The name is exact. A mistyped one returns `Unknown skill`, and an agent may
carry on without the skill. The filesystem outranks this paragraph: read the `name:` line of
`~/.grok/skills/coordinator-agents/SKILL.md` (or the repository's vendored
`.agents/skills/` copy) with your file-reading tool.

## The non-delegable default

Children's output is invisible to the user until the coordinator relays it.
Child results are input to your judgment, not the answer, and a final answer
with required results still pending is a status, not a result.
