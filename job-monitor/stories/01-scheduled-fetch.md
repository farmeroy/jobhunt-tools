# Story 1: Scheduled Job Fetching

**As a job seeker, I want the tool to automatically check my target companies for new postings on a schedule**, so I don't have to manually visit career pages.

## Acceptance Criteria

- [ ] CLI command runs a full fetch cycle for all companies with a known ATS
- [ ] Greenhouse, Ashby, and Lever fetchers pull current job listings via their public APIs
- [ ] Keyword pre-filter applied before DB insert: title whitelist (engineer, developer, software, etc.) and location whitelist (remote, US-based)
- [ ] Only postings matching both title and location whitelists are stored
- [ ] Fetchers handle rate limiting and transient errors gracefully
- [ ] Concurrent fetching across companies (bounded concurrency)
- [ ] Can be triggered manually (`job-monitor run`) or via system cron
- [ ] Logs which companies were checked, postings found, and postings filtered out

## Notes

- Build on patterns from existing `ats-detect` tool (slug generation, concurrent goroutines)
- Each ATS fetcher implements a common interface
- Companies with `ats = not-found-*` are skipped in Phase 1
- The company list is curated for remote-friendly companies, so US-based postings are assumed remote-eligible
- The keyword pre-filter reduces volume before any LLM calls (Story 3), keeping API costs low
