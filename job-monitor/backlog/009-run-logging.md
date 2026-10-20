---
title: Structured run logging to daily log files
status: partial
story: "1, 3"
---

Write run results to log files instead of just stderr.

## Done

- [x] `.logs/` directory with one file per run (timestamped, e.g. `2026-10-20T091530Z.log`)
- [x] `log.SetOutput` via `io.MultiWriter` — all `log.Printf` output goes to both stderr and the log file
- [x] Classification results logged per title (`[relevant]` / `[skip]`)
- [x] Inferred target titles logged
- [x] Token usage tracked and printed (input, output, call count)

## Remaining

- [ ] Structured output format: JSON lines or structured plain text (greppable)
- [ ] Fetch phase detail: per-company fetched/filtered/new counts in log
- [ ] Cleanup: document `find .logs/ -mtime +30 -delete` or add `--prune-logs` flag
- [ ] Scan logs for anomalies: persistent 404s (bad slugs), sudden drops in posting counts
