---
title: Per-ATS concurrency limits and retry with backoff
status: partial
story: "1"
---

The fetch phase in `run.go` uses a single global semaphore (5) across all companies regardless of ATS provider. This means all Ashby detail requests (fetched per-posting for descriptions) compete with Greenhouse and Lever listing requests. Ashby's API returns 429s when too many detail requests hit it concurrently across companies.

## Done

- [x] `ashby.go`: shared detail semaphore on `Ashby` struct (channel-based)

## Remaining

- [ ] `run.go`: per-ATS semaphores keyed by ATS name (default limit 5)
- [ ] `ashby.go`: retry with backoff on 429/5xx in `fetchDescription` (respect `Retry-After` header)
- [ ] Update `fetcher.go` to initialize `Ashby` with the shared semaphore
