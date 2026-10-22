---
title: Replace Scored boolean filter with scored_on timestamp
status: todo
story: "2"
---

The `PostingFilter.Scored *bool` field is a lossy gate — it only answers "has this ever been scored?" with no way to rescore postings older than a given date. This means `--reclassify` reclassifies titles but silently skips rescoring content for postings that were previously scored.

Replace with a `scored_on` timestamp column so the content scoring query can filter by recency the same way title classification uses `first_seen`.

## Tasks

- [ ] Add `scored_on DATETIME` column to postings (migration)
- [ ] Populate `scored_on` in `SetContentScore`
- [ ] Replace `PostingFilter.Scored *bool` with a `ScoredBefore *time.Time` filter (or similar)
- [ ] When `--reclassify` is set, pass `since` to the content scoring filter so it rescores postings scored before that window
