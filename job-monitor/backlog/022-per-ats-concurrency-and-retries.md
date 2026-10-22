---
title: Per-ATS concurrency limits and retry with backoff
status: todo
story: "3"
---

The fetch phase in `run.go` uses a single global semaphore (5) across all companies regardless of ATS provider. This means all Ashby detail requests (fetched per-posting for descriptions) compete with Greenhouse and Lever listing requests. Ashby's API returns 429s when too many detail requests hit it concurrently across companies.

## Changes

- Replace the global fetch semaphore in `run.go` with per-ATS semaphores so each provider gets its own concurrency pool
- Add retry with exponential backoff on 429/5xx in Ashby's `fetchDescription` (respect `Retry-After` header when present)
- Move the Ashby detail request semaphore from a local channel to a shared one on the `Ashby` struct so all companies share the same pool

## Tasks

- [ ] `run.go`: per-ATS semaphores keyed by ATS name (default limit 5)
- [ ] `ashby.go`: shared detail semaphore on `Ashby` struct
- [ ] `ashby.go`: retry with backoff on 429/5xx in `fetchDescription`
- [ ] Update `fetcher.go` to initialize `Ashby` with the shared semaphore
- [ ] Update backlog 015 to done once descriptions actually work end-to-end
