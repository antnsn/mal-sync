# CLAUDE.md — `mal-sync`

Project memory for Claude (and other AI agents). Read this in full at the
start of every session.

> Companion docs: [`AGENTS.md`](./AGENTS.md),
> [`.github/copilot-instructions.md`](./.github/copilot-instructions.md),
> and the role-specific prompts in [`.github/agents/`](./.github/agents/).

## What this project is

`mal-sync` is a small **Go 1.22** CLI that synchronizes Grafana-stack
configuration to backends (Alertmanager → Mimir, Mimir rules → Mimir, Loki
rules → Loki). It does not call HTTP APIs directly — it shells out to
`mimirtool` and `lokitool`. It is shipped as a Docker image
(`ghcr.io/antnsn/mal-sync`) and intended for CI/CD pipelines.

```
cmd/mal-sync/main.go          flag parsing + subcommand dispatch
internal/alertmanager/sync.go alertmanager subcommand
internal/mimirrules/sync.go   mimir-rules subcommand
internal/lokirules/sync.go    loki-rules subcommand
internal/common/utils.go      ExecuteCommand / CopyFile / EnsureDir
dockerfile                    multi-stage: mimirtool base + lokitool + Go binary
.github/workflows/            CI: docker-publish.yml (build + cosign sign)
.beads/                       beads (bd) issue tracker — DO NOT hand-edit
```

Every subcommand follows the same shape: parse flags → fall back to env
vars → snapshot inputs into a per-PID temp dir under `--temp.dir` → verify
or lint → load/sync → `defer os.RemoveAll` the temp dir.

Conventions in detail: see
[`.github/copilot-instructions.md`](./.github/copilot-instructions.md).
Implementation rules: see
[`.github/agents/go-developer.md`](./.github/agents/go-developer.md).

## 🔴 MANDATORY: run `/codex:review` before every push

Every push to a remote branch — feature, `main`, or tag — must be preceded
by **`/codex:review`** on the diff being pushed, with all bug / security /
logic findings resolved.

```text
1. git add … && git commit -m "..."
2. /codex:review                       ← NOT optional
3. Fix findings (bug/security/logic) → repeat from 2
4. go fmt ./... && go vet ./... && go build ./... && go test ./...
5. git pull --rebase
6. bd dolt push
7. git push
8. git status   → "up to date with origin"
```

If `/codex:review` is unavailable, stop and ask the user. Do not push
around the gate. Full contract:
[`.github/agents/git-workflow.md`](./.github/agents/git-workflow.md).

## Build & test

```bash
go fmt ./...
go vet ./...
go build ./...
go test ./...                       # tests are TODO — add table-driven tests
docker build -t mal-sync:dev .      # filename is lowercase `dockerfile`
```

CI: `.github/workflows/docker-publish.yml`. PR builds without pushing;
`main` and `v*.*.*` tags push to `ghcr.io/antnsn/mal-sync` and sign with
cosign.

## Architecture overview

- **Sequential by design** — one subcommand per invocation, no
  goroutines. Do not introduce concurrency without a strong reason.
- **Tool-shelling, not API calls** — the only network access happens
  inside `mimirtool` / `lokitool`. Keep it that way.
- **Stdlib only** — `go.mod` has zero non-stdlib dependencies. Adding one
  requires explicit user approval.
- **Flag + env-var symmetry** — every flag has a `MALSYNC_<SUBCMD>_<FLAG>`
  env-var fallback. Flags win.

## Conventions & patterns

- `log.Printf` / `log.Fatal` only.
- Errors wrapped with `fmt.Errorf("...: %w", err)` and meaningful context.
- Always shell out via `common.ExecuteCommand`.
- Always create temp dirs with `common.EnsureDir` and `defer
  os.RemoveAll`.
- Flag names: dotted (`rules.path`). Env vars: `MALSYNC_<SUBCMD>_<FLAG>`.
- `dockerfile` filename is lowercase — leave it alone.

## Specialized agents (Claude — pick the right prompt)

| When you are… | Read |
| --- | --- |
| Implementing Go changes | [`.github/agents/go-developer.md`](./.github/agents/go-developer.md) |
| Reviewing pre-push | [`.github/agents/code-reviewer.md`](./.github/agents/code-reviewer.md) |
| Editing Dockerfile / CI | [`.github/agents/docker-maintainer.md`](./.github/agents/docker-maintainer.md) |
| Cutting a release | [`.github/agents/release-manager.md`](./.github/agents/release-manager.md) |
| Curating bd issues | [`.github/agents/beads-steward.md`](./.github/agents/beads-steward.md) |
| Branching / committing / pushing | [`.github/agents/git-workflow.md`](./.github/agents/git-workflow.md) |

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



