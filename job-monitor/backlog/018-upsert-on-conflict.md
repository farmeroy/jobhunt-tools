---
title: UpsertPostings — use SQLite ON CONFLICT
status: todo
story: "2"
---

`UpsertPostings` currently does N individual SELECT + INSERT/UPDATE roundtrips inside a transaction. SQLite supports `INSERT INTO ... ON CONFLICT(company_id, external_id) DO UPDATE SET ...`, which collapses each row to a single statement and removes the `ErrNoRows` branching.

## Tasks

- [ ] Rewrite `UpsertPostings` to use `INSERT ... ON CONFLICT DO UPDATE`
- [ ] Verify UNIQUE constraint on `(company_id, external_id)` exists (add migration if not)
- [ ] Test with new and existing postings
