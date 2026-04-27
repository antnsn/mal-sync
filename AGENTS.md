# Agent Instructions — `mal-sync`

This file applies to **every** AI agent working in this repository
(Copilot CLI, Claude, Codex, Cursor, etc.). Read in full before acting.

> Companion docs:
> - [`CLAUDE.md`](./CLAUDE.md) — project memory & architecture
> - [`.github/copilot-instructions.md`](./.github/copilot-instructions.md) — full Copilot rules
> - [`.github/agents/`](./.github/agents/) — role-specific prompts

## 🔴 Mandatory pre-push gate: `/codex:review`

**Every push to a remote — feature branch, `main`, or tag — must be
preceded by `/codex:review` on the diff being pushed, with all
bug/security/logic findings resolved.** No exceptions. If
`/codex:review` is unavailable, stop and ask the user — do not push
around the gate. Full contract in
[`.github/agents/git-workflow.md`](./.github/agents/git-workflow.md).

## Quick context

- Go 1.22 CLI, stdlib only. Module: `github.com/antnsn/mal-sync`.
- Three subcommands (`alertmanager`, `mimir-rules`, `loki-rules`) that
  shell out to `mimirtool` / `lokitool` via `internal/common.ExecuteCommand`.
- Shipped as a Docker image `ghcr.io/antnsn/mal-sync` built and signed by
  `.github/workflows/docker-publish.yml`.
- Issue tracking: **[gastownhall/beads](https://github.com/gastownhall/beads)**
  (`bd`). No `TODO.md`, no TodoWrite, no markdown task lists.

## Beads (`bd`) — quick reference

```bash
bd prime                # Read first — full workflow context
bd ready                # Find available work
bd show <id>            # View issue details
bd update <id> --claim  # Claim work atomically
bd close <id>           # Complete work
bd dep add <child> <parent>
bd remember "..."       # Persistent project knowledge
bd dolt push            # Push beads data to remote
```

## Non-Interactive Shell Commands

**ALWAYS use non-interactive flags** with file operations to avoid hanging on confirmation prompts.

Shell commands like `cp`, `mv`, and `rm` may be aliased to include `-i` (interactive) mode on some systems, causing the agent to hang indefinitely waiting for y/n input.

**Use these forms instead:**
```bash
# Force overwrite without prompting
cp -f source dest           # NOT: cp source dest
mv -f source dest           # NOT: mv source dest
rm -f file                  # NOT: rm file

# For recursive operations
rm -rf directory            # NOT: rm -r directory
cp -rf source dest          # NOT: cp -r source dest
```

**Other commands that may prompt:**
- `scp` - use `-o BatchMode=yes` for non-interactive
- `ssh` - use `-o BatchMode=yes` to fail instead of prompting
- `apt-get` - use `-y` flag
- `brew` - use `HOMEBREW_NO_AUTO_UPDATE=1` env var

<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:ca08a54f -->
## Beads Issue Tracker

This project uses **bd (beads)** for issue tracking. Run `bd prime` to see full workflow context and commands.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --claim  # Claim work
bd close <id>         # Complete work
```

### Rules

- Use `bd` for ALL task tracking — do NOT use TodoWrite, TaskCreate, or markdown TODO lists
- Run `bd prime` for detailed command reference and session close protocol
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files

## Session Completion

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd dolt push
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds
<!-- END BEADS INTEGRATION -->
