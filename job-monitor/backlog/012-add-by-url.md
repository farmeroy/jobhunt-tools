---
title: Add company with validation fetch
status: todo
story: "1"
---

Improve `add` to validate the company by fetching jobs before saving. Optionally accept a career page URL directly.

## Behavior

- Detect ATS (or parse `--url` if provided) → fetch jobs → if fetch succeeds, save company + upsert postings → if fetch fails, error and don't save
- Fail fast: no more silently adding companies with bad slugs that get skipped during `run`
- `--url` flag parses known ATS URL patterns (e.g. `https://jobs.lever.co/nerdwallet` → lever + nerdwallet), skipping probe-based detection

## Tasks

- [ ] URL parser for known ATS patterns (greenhouse, lever, ashby)
- [ ] `--url` flag on `add` command
- [ ] Validate by fetching: only persist company if fetch succeeds
- [ ] Upsert fetched postings on successful add (first fetch for free)
- [ ] Fall back to existing DetectATS when no URL given
