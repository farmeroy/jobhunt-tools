---
title: ATS fetchers (Greenhouse, Ashby, Lever)
status: todo
story: "1"
---

Implement fetchers for the three Phase 1 ATS providers behind a common `Fetcher` interface.

## Tasks

- [ ] `Fetcher` interface — `Name() string`, `Fetch(ctx, slug) ([]JobPosting, error)`
- [ ] Greenhouse fetcher — `boards-api.greenhouse.io/v1/boards/{slug}/jobs?content=true`
- [ ] Ashby fetcher — `jobs.ashbyhq.com/api/non-user-graphql` (GraphQL)
- [ ] Lever fetcher — `api.lever.co/v0/postings/{slug}`
- [ ] Shared HTTP client with retry on 429/5xx
- [ ] Slug resolution — try variants, cache working slug back to DB
