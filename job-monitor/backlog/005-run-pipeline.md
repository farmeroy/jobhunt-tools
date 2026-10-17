---
title: Run command — fetch + dedup pipeline
status: done
story: "1, 2"
---

Wire up `job-monitor run` to orchestrate the fetch-and-dedup pipeline.

## Done

- `cmd/run.go` — `job-monitor run` with bounded concurrency (5 goroutines), `--dry-run` flag
- `fetcher/prefilter.go` — keyword pre-filter: title whitelist + location whitelist (regex word boundary for "US")
- `fetcher/prefilter_test.go` — 3 test functions covering title matching, location matching, and combined filter
- Fetchers updated to return `Location` field (Greenhouse: `location.name`, Lever: `categories.location`, Ashby: `locationName`)
- Pipeline: fetch → pre-filter → upsert → mark stale → log summary
- Dry-run tested: 3021 fetched → 862 matched filter across 66 companies (71% reduction)
