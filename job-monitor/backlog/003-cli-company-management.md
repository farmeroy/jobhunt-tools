---
title: CLI commands for company management
status: done
story: "6"
---

Build the CLI subcommands that wire up the store's company CRUD to user-facing commands.

## Done

- `main.go` — cobra entry point
- `cmd/root.go` — root command, `--db` flag, store lifecycle (open/close)
- `cmd/import.go` — `job-monitor import <csv>` seeds companies from CSV, skips `not-found-*` rows, skips duplicates (UNIQUE name)
- `cmd/add.go` — `job-monitor add <company>` with auto-detect ATS via `atsutil.DetectATS()`
- `cmd/remove.go` — `job-monitor remove <company>` soft-delete (deactivate)
- `cmd/list.go` — `job-monitor list` with `--all` flag for inactive
- `cmd/cmd_test.go` — 6 tests (import, duplicate skip, remove, list empty/active/all)
- Schema: added `UNIQUE(name)` to companies table
- Upsert/update on reimport deferred to issue 007
