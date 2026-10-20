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

- **Distinct titles, not individual postings** — `GetDistinctTitles()` deduplicates before sending to the LLM, then `SetTitleRelevanceByTitle()` bulk-updates all postings with that title
- **Two-step prompt** — first call infers target titles from the resume, second call classifies against those targets. Keeps the classification prompt focused and cheaper (no resume in the classify call)
- **Resume text via pdftotext** — `resume.LoadText(path)` extracts text from PDF via `pdftotext` (poppler-utils), also supports .txt/.md
- Uses Anthropic tool_use for structured output — guaranteed JSON schema
- Batching: titles chunked at 100 per API call, numeric indices for robust round-tripping
- `--reclassify` flag re-runs classification on postings within a time period (e.g. `--reclassify 7d`)
- `run` command is the full pipeline: fetch → classify. `--resume` flag triggers classification; without it, only fetch runs
- A keyword pre-filter (title whitelist + location whitelist) already runs at fetch time (Story 1) before postings reach the DB
- Token usage tracked per run (input, output, call count)
- All classification results logged to `.logs/` with per-title detail

## Known Issues

- Title filter is too strict with the current prompt — needs loosening to be a wide net (content filter should be the strict pass)
- Seniority inference from resume is imprecise — blacklist regex for explicit exclusions is a backlog item
