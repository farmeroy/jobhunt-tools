---
title: Migrations and query layer
status: partial
story: "2"
---

Replace the handrolled store layer with versioned migrations and a proper query tool.

## Migrations — done

Replaced the single `CREATE TABLE IF NOT EXISTS` with goose (`pressly/goose/v3`). SQL migration files live in `store/*.sql`, embedded via `//go:embed`. `NewProvider` API (concurrency-safe, no global state). Existing databases pick up migration 001 idempotently.

- [x] Pick library → goose
- [x] Convert existing schema to migration 001
- [x] Run migrations on Open()

## Query layer

Replace raw `database/sql` string-building with a typed query approach.

Options to evaluate:
- **sqlc** — write SQL, generates type-safe Go code (popular in Go, no runtime overhead)
- **sqlx** — thin wrapper, struct scanning + named params (familiar from Rust ecosystem)
- **squirrel** — query builder, replaces string concat

- [ ] Pick library
- [ ] Replace queryPostings / manual Scan with generated or struct-mapped queries
- [ ] Replace manual placeholder building (MarkStale, MarkNotified IN clauses)
- [ ] Remove `WHERE 1=1` dynamic query pattern
