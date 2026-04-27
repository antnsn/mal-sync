# GitHub Copilot Instructions — `mal-sync`

These instructions apply to **all** Copilot surfaces (Copilot CLI, Copilot Chat in
VS Code / JetBrains / GitHub.com, Copilot code completion, and Copilot coding
agents) when working in this repository.

> See also: [`CLAUDE.md`](../CLAUDE.md), [`AGENTS.md`](../AGENTS.md), and
> [`.github/agents/`](./agents/) for role-specific agent prompts.

---

## 1. Project context (deep dive)

`mal-sync` is a small **Go 1.22** CLI that synchronizes Grafana-stack
configuration (Alertmanager, Mimir rules, Loki rules) to the corresponding
backends. It is intended to be run inside a container in CI/CD pipelines. It
does not call HTTP APIs directly — it shells out to `mimirtool` and
`lokitool`, which are baked into the Docker image.

### Repository layout

```
cmd/mal-sync/main.go          # Entry point: flag parsing + subcommand dispatch
internal/alertmanager/sync.go # `alertmanager` subcommand: copy → verify → load via mimirtool
internal/mimirrules/sync.go   # `mimir-rules` subcommand: copy → lint → sync via mimirtool
internal/lokirules/sync.go    # `loki-rules`  subcommand: copy → lint → sync via lokitool
internal/common/utils.go      # ExecuteCommand, CopyFile, EnsureDir helpers
dockerfile                    # Multi-stage build: mimirtool base + lokitool + Go binary
.github/workflows/            # CI: docker-publish.yml builds & signs images to GHCR
.beads/                       # Beads (bd) issue tracker state — DO NOT hand-edit
```

### Subcommand contract (each subcommand follows this pattern)

1. Parse flags, then fall back to `MALSYNC_<CMD>_<FLAG>` env vars (flags win).
2. Validate required inputs and `log.Fatal` with a clear message if missing.
3. Create a per-PID temp directory under `--temp.dir` (default `/tmp`) and
   `defer os.RemoveAll`.
4. Snapshot input files into the temp dir (so mutating sources mid-sync is
   safe).
5. Verify/lint each file with the upstream tool.
6. Load/sync to the backend.

When **adding a subcommand**, replicate this shape and add a section to
`README.md` with both the flag and env-var names.

### Conventions

- Module path: `github.com/antnsn/mal-sync`. New packages go under `internal/`.
- Use `log.Printf` / `log.Fatal` (the existing style); do not pull in a
  logging framework unless asked.
- Always use `common.ExecuteCommand` to shell out — it already logs the
  command line and combined output on failure.
- Always use `common.EnsureDir` and `common.CopyFile` rather than
  reimplementing them.
- Errors must be wrapped with `fmt.Errorf("...: %w", err)` and include
  enough context to identify the failing file/path.
- `defer os.RemoveAll(syncTempDir)` is mandatory after any temp-dir creation.
- Flag names use dotted form (`rules.path`, `mimir.address`); env vars use
  `MALSYNC_<SUBCMD>_<FLAG_UPPER_SNAKE>`.
- No non-stdlib dependencies have been introduced. Do not add one without
  explicit approval — keep `go.mod` minimal.

---

## 2. Build, lint, test

```bash
go fmt ./...
go vet ./...
go build ./...
go test ./...                  # tests are TODO — add table-driven tests next to code
docker build -t mal-sync:dev . # uses lowercase `dockerfile`; requires -f if your Docker is strict
```

The CI workflow `.github/workflows/docker-publish.yml` builds and signs the
image with cosign. PRs build but do not push. Tags `v*.*.*` and pushes to
`main` publish to `ghcr.io/antnsn/mal-sync`.

---

## 3. Issue tracking — Beads (`bd`)

This project uses **[gastownhall/beads](https://github.com/gastownhall/beads)**
for **all** task tracking. `.beads/` is committed to the repo.

**Do NOT** create markdown TODO lists, `TODO.md`, or use any other tracker.
Use `bd` exclusively.

```bash
bd prime                  # Read this first — full workflow context
bd ready                  # List unblocked tasks
bd create "Title" -p 1    # Create P1 task
bd update <id> --claim    # Atomically claim a task
bd show <id>              # View details + audit trail
bd close <id>             # Mark done
bd dep add <child> <parent>
bd remember "..."         # Persistent project knowledge (replaces MEMORY.md)
bd dolt push              # Push beads data to remote (Dolt)
```

When you finish a unit of work, file follow-up issues with `bd create` for
anything you discovered but did not address.

---

## 4. **MANDATORY pre-push gate: `/codex:review`**

> 🔴 **You MUST run `/codex:review` and resolve its findings before any
> `git push` to a remote branch — including feature branches, `main`, and
> tags.** This is non-negotiable.

`/codex:review` is the Codex Claude Code plugin's review slash-command. It
runs an independent review of the staged/branch diff and surfaces real
correctness, security, and logic issues (it does not nit on style).

### Workflow before every push

```text
1. git add ... && git commit -m "..."
2. /codex:review                        ← run BEFORE pushing
3. Triage findings:
     - Bug / security / logic error → fix in a follow-up commit, repeat from 2
     - Style / preference / out-of-scope → file a `bd create` follow-up if useful, otherwise dismiss
4. Run quality gates: `go fmt ./...`, `go vet ./...`, `go build ./...`, `go test ./...`
5. git pull --rebase
6. bd dolt push
7. git push
8. git status                           ← MUST show "up to date with origin"
```

If Codex is unavailable, escalate to the user — do not push around the gate.
A `pre-push` git hook (see [`.github/agents/git-workflow.md`](./agents/git-workflow.md))
documents the same contract for human contributors.

### What does **not** require `/codex:review`

- Local commits that you have not pushed.
- Edits to session-scratch files outside the repo.

Everything else does.

---

## 5. Pull request expectations

- Title: imperative mood, ≤ 72 chars (e.g. `loki-rules: respect --temp.dir for staging`).
- Body: link the bd issue (`Closes bd-XXXX`), summarize the change, list
  manual verification steps. The CI image build must be green.
- Keep PRs scoped. Refactors and feature work live in separate PRs.
- Update `README.md` whenever you add/rename a flag or env var.

---

## 6. Specialized agents

Role-specific prompts live in [`.github/agents/`](./agents/). Invoke them
when the task matches their scope:

| Agent file | Use when… |
| --- | --- |
| [`go-developer.md`](./agents/go-developer.md) | Adding/modifying Go code under `cmd/` or `internal/` |
| [`code-reviewer.md`](./agents/code-reviewer.md) | Performing pre-push review (mirrors `/codex:review` rubric) |
| [`docker-maintainer.md`](./agents/docker-maintainer.md) | Touching `dockerfile` or the publish workflow |
| [`release-manager.md`](./agents/release-manager.md) | Cutting tags, updating release notes |
| [`beads-steward.md`](./agents/beads-steward.md) | Curating bd issues, dependencies, compaction |
| [`git-workflow.md`](./agents/git-workflow.md) | Branching, commits, the pre-push contract |

---

## 7. Shell hygiene

Some shells alias `cp`/`mv`/`rm` to interactive mode. Always use the
non-interactive forms in scripted contexts:

```bash
cp -f src dst       # not: cp src dst
mv -f src dst
rm -f file
rm -rf directory
ssh -o BatchMode=yes ...
scp -o BatchMode=yes ...
apt-get -y ...
HOMEBREW_NO_AUTO_UPDATE=1 brew ...
```

---

## 8. Things to refuse / escalate

- Adding new third-party Go dependencies without explicit approval.
- Committing secrets, kubeconfigs, or real Mimir/Loki tenant IDs.
- Pushing without `/codex:review`.
- Editing `.beads/embeddeddolt/` by hand — always go through `bd`.
