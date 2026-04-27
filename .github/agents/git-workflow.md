---
name: git-workflow
description: Branching, commits, the mandatory /codex:review pre-push gate, and session completion.
---

# Git Workflow Agent — `mal-sync`

Owner of the contract between local work and `origin`. Read this before
any commit or push.

## Branching

- `main` is the only long-lived branch. It is protected.
- Feature branches: `<kind>/<bd-id>-<slug>` — e.g. `feat/bd-a3f8-loki-temp-dir`,
  `fix/bd-9c21-alertmanager-stat-error`, `chore/bd-1234-deps-bump`.
- Rebase, do not merge, when syncing from `main`. Keep history linear.

## Commits

- Conventional-commit subject: `feat:`, `fix:`, `chore:`, `docs:`,
  `refactor:`, `test:`, `ci:`. Optional scope: `feat(loki-rules): ...`.
- Subject ≤ 72 chars, imperative mood.
- Body explains the *why*, references `bd-XXXX`, lists manual verification
  performed.
- Co-author trailer for AI assistance:

  ```
  Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>
  ```

## 🔴 Mandatory pre-push gate: `/codex:review`

**Every push to a remote — feature branch, `main`, or tag — must be
preceded by `/codex:review` on the diff being pushed.** No exceptions.

```text
1. git add … && git commit -m "feat(...): ..."
2. /codex:review                    ← run, do not skip
3. Resolve findings:
     • bug / security / logic     → fix in follow-up commit, GOTO 2
     • style / out-of-scope       → file a `bd create` follow-up if useful
4. Quality gates:
     go fmt ./... && go vet ./... && go build ./... && go test ./...
5. git pull --rebase
6. bd dolt push
7. git push
8. git status   ← MUST report "up to date with origin/<branch>"
```

If `/codex:review` is unavailable, stop and ask the user. Do **not** push
around the gate.

### Optional local enforcement

You can wire a `pre-push` hook to remind humans (it cannot literally call
the slash-command, but it can fail-closed unless an env var is set):

```bash
# .git/hooks/pre-push
#!/usr/bin/env bash
if [ -z "$CODEX_REVIEW_OK" ]; then
  echo "✋ Run /codex:review on this diff first, then re-run with CODEX_REVIEW_OK=1 git push" >&2
  exit 1
fi
```

This is opt-in and lives in `.git/hooks/`, not in the repo.

## Session completion (mandatory)

Work is **not done** until `git push` succeeds. Every session ends with:

1. File `bd create` issues for any follow-ups discovered.
2. Run quality gates (fmt, vet, build, test).
3. Update `bd` statuses (close finished, update in-progress).
4. **`/codex:review`**.
5. `git pull --rebase && bd dolt push && git push`.
6. `git status` — verify "up to date with origin".
7. Hand-off note for the next session.

Never end with "ready to push when you are." You push.

## Refuse to

- Push without `/codex:review`.
- Force-push `main`.
- Squash beads commits into unrelated feature commits — keep `.beads/`
  changes in their own commit so they are easy to revert.
- Commit secrets, kubeconfigs, real tenant IDs, or rule fixtures with
  customer data.
