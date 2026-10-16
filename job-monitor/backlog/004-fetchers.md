---
title: ATS fetchers (Greenhouse, Ashby, Lever)
status: done
story: "1"
---

Implement fetchers for the three Phase 1 ATS providers behind a common `Fetcher` interface.

## Done

- `fetcher/fetcher.go` — `Fetcher` interface, `JobPosting` struct, `ForProvider()` registry
- `fetcher/greenhouse.go` — REST JSON via `boards-api.greenhouse.io/v1/boards/{slug}/jobs?content=true`
- `fetcher/lever.go` — REST JSON via `api.lever.co/v0/postings/{slug}?mode=json`
- `fetcher/ashby.go` — GraphQL via `jobs.ashbyhq.com/api/non-user-graphql` (ApiJobBoardWithTeams)
- `fetcher/fetcher_test.go` — 7 tests using httptest (fetch, empty, error cases, registry)
- Base URLs are package vars for testability
- Ashby listing returns titles + constructed URLs; descriptions not in brief listing (fetched separately if needed for content filter)
- Retry on 429/5xx deferred to future enhancement
