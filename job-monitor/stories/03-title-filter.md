# Story 3: LLM Title Filter (Pass 1)

**As a job seeker, I want new postings filtered by job title first**, so irrelevant roles are discarded cheaply before deeper analysis.

## Acceptance Criteria

- [ ] New postings are sent through a title relevance check using Haiku 4.5
- [ ] Multiple titles are batched into a single API call for efficiency
- [ ] Each title gets a binary relevant/irrelevant classification
- [ ] Classification result is stored in the database (`title_relevant` field)
- [ ] User's resume is included in the prompt as context for relevance judgment
- [ ] Simple string pre-filter skips obviously wrong seniority (intern, staff, principal, VP, director) before hitting the LLM
- [ ] Irrelevant postings are logged but not processed further

## Notes

- Resume is loaded once and included in the system prompt
- Batching: send all new titles from one fetch cycle in a single call (or chunked if too many)
- Cost estimate: ~1 call per fetch cycle, minimal token usage
- The pre-filter is just string matching to save API calls on obvious mismatches
