---
title: Run command — fetch + dedup pipeline
status: todo
story: "1, 2"
---

Wire up `job-monitor run` to orchestrate the fetch-and-dedup pipeline.

## Tasks

- [ ] `job-monitor run` command — fetch all active companies with bounded concurrency
- [ ] Call fetcher per company, upsert postings, mark stale
- [ ] Log summary: companies checked, postings found, new postings
- [ ] `--dry-run` flag — fetch but don't persist
