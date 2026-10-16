---
title: SQLite store layer with companies and postings tables
status: done
story: "2, 6"
---

Set up the SQLite persistence layer with schema, models, and CRUD operations for companies and postings.

## Done

- `store/store.go` — `Open()`, `Close()`, auto-migration
- `store/migrations.go` — DDL for `companies` and `postings` tables
- `store/models.go` — `Company` and `Posting` structs
- `store/company.go` — `AddCompany`, `DeactivateCompany`, `ListActiveCompanies`, `ListAllCompanies`, `UpdateSlug`
- `store/posting.go` — `UpsertPostings`, `MarkStale`, `GetNewUnfilteredPostings`, `GetTitlePassedUnscored`, `GetUnnotifiedMatches`, `SetTitleRelevance`, `SetContentScore`, `MarkNotified`
- `store/store_test.go` — 11 tests using in-memory SQLite
- `notified_at` and `applied_at` are nullable timestamps (not booleans)
- `MarkStale` handles empty board case (all postings marked stale)
