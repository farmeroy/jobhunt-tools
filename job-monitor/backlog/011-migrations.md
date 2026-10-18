---
title: Migrations and query layer
status: todo
story: "2"
---

Replace the handrolled store layer with versioned migrations and a proper query tool.

## Migrations

Replace the single CREATE TABLE IF NOT EXISTS with versioned migrations using goose or golang-migrate.

- [ ] Pick library (goose or golang-migrate)
- [ ] Convert existing schema to migration 001
- [ ] Run migrations on Open()

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
