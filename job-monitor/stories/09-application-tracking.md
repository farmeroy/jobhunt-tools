# Story 9 (Stretch): Application Tracking

**As a job seeker, I want to track which matched postings I applied to vs. skipped**, closing the loop between discovery and action.

## Acceptance Criteria

- [ ] `job-monitor apply <posting-id>` marks a posting as applied
- [ ] `job-monitor skip <posting-id>` marks a posting as skipped (with optional reason)
- [ ] `job-monitor status` shows a summary: matched, applied, skipped, pending
- [ ] Applied/skipped postings are excluded from future notifications

## Notes

- Lightweight — not trying to replace a full ATS/CRM, just closing the feedback loop
- Could eventually feed back into the LLM filter to improve relevance over time
