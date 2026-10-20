---
title: Store filter struct — replace single-purpose query methods
status: todo
story: "2"
---

Replace `GetNewUnfilteredPostings`, `GetTitlePassedUnscored`, `GetUnnotifiedMatches`, and `GetDistinctTitles` with a single `ListPostings(PostingFilter)` method that composes WHERE clauses from optional filter fields.

## Motivation

Each pipeline stage added a new bespoke query method, duplicating the full 14-column SELECT list and `queryPostings` call pattern. This also caused a bug: the `postings --min-score` flag dispatches to `GetUnnotifiedMatches`, which silently adds `notified_at IS NULL` — hiding notified postings the user asked to see.

A filter struct makes the store testable, composable, and prevents CLI commands from accidentally inheriting pipeline-specific WHERE clauses.

## Tasks

- [ ] Define `PostingFilter` struct with optional fields: `TitleRelevant *bool`, `ContentScoreMin *float64`, `Notified *bool`, `Scored *bool`, `CompanyID *int64`, `DistinctTitles bool`
- [ ] Implement `ListPostings(PostingFilter)` with dynamic WHERE clause building
- [ ] Migrate all callers (run.go pipeline stages, postings.go CLI) to use `ListPostings`
- [ ] Remove the old single-purpose methods
- [ ] Apply same pattern to `ListActiveCompanies` / `ListAllCompanies` → `ListCompanies(activeOnly *bool)`
- [ ] Fix `postings --min-score` to filter only by score, not notification status
