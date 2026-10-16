---
title: Run command — fetch + dedup pipeline
status: todo
story: "1, 2"
---

Wire up `job-monitor run` to orchestrate the fetch-and-dedup pipeline.

## Tasks

- [ ] `job-monitor run` command — fetch all active companies with bounded concurrency
- [ ] Keyword pre-filter: title whitelist (engineer, developer, software, etc.) + location whitelist (remote, US-based) — drop non-matching postings before DB insert
- [ ] Fetchers return location data alongside title/URL for the location filter
- [ ] Call fetcher per company, filter, upsert postings, mark stale
- [ ] Log summary: companies checked, postings found, postings filtered, new postings stored
- [ ] `--dry-run` flag — fetch and filter but don't persist
