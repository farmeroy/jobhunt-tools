# Story 3: LLM Title Filter (Pass 1)

**As a job seeker, I want new postings filtered by job title first**, so irrelevant roles are discarded cheaply before deeper analysis.

## Acceptance Criteria

- [ ] New postings are sent through a title relevance check using Haiku 4.5
- [ ] Multiple titles are batched into a single API call for efficiency
- [ ] Each title gets a binary relevant/irrelevant classification
- [ ] Classification result is stored in the database (`title_relevant` field)
- [ ] User's resume is included in the prompt as context for relevance judgment
- [ ] Irrelevant postings are logged but not processed further

## Notes

- Resume is loaded once and included in the system prompt
- Batching: send all new titles from one fetch cycle in a single call (or chunked if too many)
- Cost estimate: ~1 call per fetch cycle, minimal token usage
- A keyword pre-filter (title whitelist + location whitelist) already runs at fetch time (Story 1) before postings reach the DB — the LLM only sees postings that survived that filter
- The LLM handles edge cases the keyword filter can't: ambiguous titles, seniority level judgment, relevance to the user's specific background
