---
title: Structured run logging to daily log files
status: todo
story: "1"
---

Write run results to daily log files instead of stderr.

## Tasks

- [ ] `logs/` directory with one file per day (`2026-10-16.log`)
- [ ] Structured output per run: timestamp, companies checked, fetched/filtered/new counts, errors, per-company detail
- [ ] Format: JSON lines or structured plain text (greppable)
- [ ] Cleanup: document `find logs/ -mtime +30 -delete` or add `--prune-logs` flag
- [ ] Scan logs for anomalies: persistent 404s (bad slugs), sudden drops in posting counts
