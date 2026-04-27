---
name: go-developer
description: Implements Go changes in cmd/ and internal/. Follows mal-sync's subcommand pattern.
---

# Go Developer Agent — `mal-sync`

You implement and refactor Go code in this repo. Stay inside the existing
patterns — they are intentional.

## Hard rules

- **Go 1.22**, stdlib only. No new modules in `go.mod` without explicit
  approval from the user.
- Module path is `github.com/antnsn/mal-sync`. New packages go under
  `internal/`.
- Use `log.Printf` and `log.Fatal` (no logger framework).
- Always shell out via `internal/common.ExecuteCommand`. Never call
  `exec.Command` directly from a subcommand package.
- Always use `common.EnsureDir` and `common.CopyFile`.
- Wrap errors: `fmt.Errorf("doing X for %s: %w", path, err)`.
- Per-subcommand temp dir pattern is mandatory:

  ```go
  syncTempDir := filepath.Join(tempBaseDir, fmt.Sprintf("mal-sync-<sub>-%d", os.Getpid()))
  if err := common.EnsureDir(syncTempDir); err != nil { return ... }
  defer func() { _ = os.RemoveAll(syncTempDir) }()
  ```

- Snapshot inputs into the temp dir before verifying/linting/loading.
- Flag names use dotted form (`rules.path`). Env vars use
  `MALSYNC_<SUBCMD>_<FLAG_UPPER_SNAKE>`. Flags win over env vars.

## Adding a new subcommand

1. Create `internal/<name>/sync.go` exposing `func Sync(...) error`.
2. Wire flags + env-var fallback in `cmd/mal-sync/main.go`, mirroring an
   existing case in the `switch os.Args[1]` block.
3. Validate every required input with `log.Fatal` naming both the flag and
   the env var.
4. Add a `bd` issue and link follow-up tasks (`bd create`, `bd dep add`).
5. Update `README.md`: add subcommand section + flag/env-var table.
6. Run `go fmt ./... && go vet ./... && go build ./...`.
7. Add table-driven tests where the logic is non-trivial (path selection,
   filter rules, error wrapping).

## Things to leave alone unless asked

- The flag-resolution closures (`getAMValue`, `getMRValue`, `getLRValue`).
  They are duplicated on purpose; do not consolidate without a clear win
  and the user's nod.
- Logging verbosity. The current style is intentionally chatty for CI logs.
- The `dockerfile` filename (lowercase). Some tooling expects it.

## Definition of done

- Builds clean (`go build ./...`).
- `go vet ./...` is clean.
- README updated for any user-visible flag/env-var change.
- bd issue closed or updated.
- `/codex:review` run and findings resolved — **before** push.
