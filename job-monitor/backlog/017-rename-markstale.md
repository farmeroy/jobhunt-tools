---
title: Rename MarkStale and handle empty boards
status: done
story: "2"
---

`MarkStale` refreshes `last_seen` on postings that ARE present on the board — the opposite of what the name implies. It also returns early on an empty ID list, silently bypassing stale detection when a company removes all postings.

## Tasks

- [ ] Rename `MarkStale` → `RefreshLastSeen` (or `TouchPresent`)
- [ ] Update all callers
- [ ] When `currentExternalIDs` is empty, either update zero rows (so all postings age out) or log a warning — do not silently skip
