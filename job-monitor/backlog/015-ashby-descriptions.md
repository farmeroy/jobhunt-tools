---
title: Ashby fetcher missing job descriptions
status: done
story: "4"
---

The Ashby fetcher's GraphQL query only requested `id`, `title`, and `locationName` — no description field.

## Done

- [x] Ashby `fetchDescription` makes a per-posting GraphQL call requesting `descriptionHtml`
- [x] Descriptions populated during fetch via a detail request per posting
- [x] Existing postings backfilled on next `run` (re-fetched via upsert)
