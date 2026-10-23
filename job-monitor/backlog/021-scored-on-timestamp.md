---
title: Replace Scored boolean filter with scored_on timestamp
status: done
story: "2"
---

The `PostingFilter.Scored *bool` field is a lossy gate — it only answers "has this ever been scored?" with no way to rescore postings older than a given date. This means `--reclassify` reclassifies titles but silently skips rescoring content for postings that were previously scored.

Replace with a `scored_on` timestamp column so the content scoring query can filter by recency the same way title classification uses `first_seen`.

## Tasks

- [x] Add `scored_on DATETIME` column to postings (migration)
- [x] Populate `scored_on` in `SetContentScore`
- [x] Replace `PostingFilter.Scored *bool` with `Unscored bool` + `ScoredSince *time.Time`
- [x] When `--reclassify` is set, pass `since` to the content scoring filter so it rescores postings scored within that window
- [x] Clear `notified_at` on rescored postings so they flow through to matches again

## Commits

- 303096d feat: replace Scored boolean with scored_on timestamp for reclassify support
