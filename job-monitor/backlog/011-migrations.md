---
title: Proper schema migration system
status: todo
story: "2"
---

Replace the single CREATE TABLE IF NOT EXISTS with versioned migrations using goose or golang-migrate.

## Tasks

- [ ] Pick library (goose or golang-migrate)
- [ ] Convert existing schema to migration 001
- [ ] Add migration 002 for any future schema changes
- [ ] Run migrations on Open()
