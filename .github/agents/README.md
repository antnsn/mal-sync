# Agents — `mal-sync`

This directory holds role-specific agent prompts. Any AI agent (Copilot,
Claude, Codex) working in this repo should pick the prompt that matches
the task at hand and follow it in addition to:

- [`/.github/copilot-instructions.md`](../copilot-instructions.md) — top-level Copilot rules
- [`/CLAUDE.md`](../../CLAUDE.md) — Claude project memory
- [`/AGENTS.md`](../../AGENTS.md) — generic agent rules (beads, shell hygiene, etc.)

## Roster

| Agent | Scope |
| --- | --- |
| [`go-developer`](./go-developer.md) | Implement / refactor Go code in `cmd/` and `internal/` |
| [`code-reviewer`](./code-reviewer.md) | Pre-push review (mirrors `/codex:review` rubric) |
| [`docker-maintainer`](./docker-maintainer.md) | `dockerfile` + `docker-publish.yml` |
| [`release-manager`](./release-manager.md) | Tagging, release notes, image verification |
| [`beads-steward`](./beads-steward.md) | `bd` graph hygiene, compaction, dependencies |
| [`git-workflow`](./git-workflow.md) | Branching, commits, the **`/codex:review` pre-push gate** |

## The non-negotiable

> Every push to a remote — feature branch, `main`, or tag — must be
> preceded by **`/codex:review`** on the diff being pushed. See
> [`git-workflow.md`](./git-workflow.md) for the full sequence.
