# Story 3: LLM Title Filter (Pass 1)

**As a job seeker, I want new postings filtered by job title first**, so irrelevant roles are discarded cheaply before deeper analysis.

## Acceptance Criteria

- [x] New postings are sent through a title relevance check using Haiku 4.5
- [x] Multiple titles are batched into a single API call for efficiency (chunked at 100)
- [x] Each title gets a binary relevant/irrelevant classification
- [x] Classification result is stored in the database (`title_relevant` field)
- [x] User's resume is included in the prompt as context for relevance judgment
- [x] Irrelevant postings are logged but not processed further

## Implementation Notes

- Resume is loaded once via `resume.LoadText(path)` and included in the system prompt
- Uses Anthropic tool_use for structured output — guaranteed JSON schema
- Batching: titles chunked at 100 per API call
- Classification runs on **all unclassified postings** (`title_relevant IS NULL`), not just the current run's fetch — handles backfill from failed runs, newly added companies, DB re-seeds
- Resume updates don't reclassify old postings — only new unclassified ones get the updated context
- `run` command is the full pipeline: fetch → classify. `--resume` flag triggers classification; without it, only fetch runs
- A keyword pre-filter (title whitelist + location whitelist) already runs at fetch time (Story 1) before postings reach the DB — the LLM only sees postings that survived that filter
- The LLM handles edge cases the keyword filter can't: ambiguous titles, seniority level judgment, relevance to the user's specific background
- Prompt eval to be built from real classification mistakes after initial results
