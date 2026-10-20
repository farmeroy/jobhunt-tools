---
title: Ashby fetcher missing job descriptions
status: todo
story: "4"
---

The Ashby fetcher's GraphQL query only requests `id`, `title`, and `locationName` — no description field. All 36 Ashby postings that passed the title filter have empty descriptions, which blocks the content scoring pass.

Greenhouse and Lever both populate descriptions correctly.

## Tasks

- [ ] Investigate Ashby GraphQL schema for description fields (likely `descriptionHtml` or `descriptionPlain`)
- [ ] Add description to the Ashby query
- [ ] Backfill existing Ashby postings (re-fetch or separate update query)
