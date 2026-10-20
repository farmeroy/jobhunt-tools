---
title: Title blacklist regex for seniority exclusion
status: todo
story: "3"
---

Add a regex blacklist to exclude titles by seniority level (e.g. Staff, Principal, Distinguished, Director, VP). Similar to the existing title whitelist in `fetcher/prefilter.go` but inverted.

This is a deterministic pre-filter — runs before the LLM, no API cost.

## Tasks

- [ ] Add blacklist patterns to `fetcher/prefilter.go` alongside existing whitelist
- [ ] Default patterns: Staff, Principal, Distinguished, Director, VP (configurable later)
- [ ] Apply at the same stage as the existing keyword pre-filter
