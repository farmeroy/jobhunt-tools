---
title: Run command — full pipeline (fetch + classify)
status: done
story: "1, 2, 3"
---

Wire up `job-monitor run` to orchestrate the full pipeline: fetch → classify.

## Done

- `cmd/run.go` — `job-monitor run` with bounded concurrency (5 goroutines), `--dry-run` flag, `--resume` flag
- `fetcher/prefilter.go` — keyword pre-filter: title whitelist + location whitelist (regex word boundary for "US")
- `fetcher/prefilter_test.go` — 3 test functions covering title matching, location matching, and combined filter
- Fetchers updated to return `Location` field (Greenhouse: `location.name`, Lever: `categories.location`, Ashby: `locationName`)
- Pipeline: fetch → pre-filter → upsert → mark stale → LLM title classify → log summary
- `filter/title.go` — two-step title classification via Haiku 4.5: infer targets from resume, then classify distinct titles. Tool_use for structured output, chunks at 100
- `resume/resume.go` — PDF text extraction via pdftotext (poppler-utils), also supports .txt/.md
- Classification runs on all distinct titles; `--reclassify` flag for re-running within a time window
- Token usage tracked and logged per run
- Run logs persisted to `.logs/` directory (timestamped per run)
- Dry-run tested: 3021 fetched → 862 matched filter across 66 companies (71% reduction)
