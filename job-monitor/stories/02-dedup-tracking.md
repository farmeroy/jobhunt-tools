# Story 2: Dedup & New Post Detection

**As a job seeker, I want the tool to track which postings it has already seen**, so I only get notified about genuinely new opportunities.

## Acceptance Criteria

- [x] SQLite database stores all fetched postings with a unique key (company + ATS job ID)
- [x] On each fetch cycle, only postings not already in the database are flagged as "new"
- [x] Previously seen postings are updated if their title or description changed
- [x] Postings that disappear from a company's board are marked as closed/removed (not deleted)
- [x] `first_seen` timestamp recorded for every posting

## Notes

- The dedup key needs to be ATS-specific — Greenhouse uses numeric IDs, Ashby uses UUIDs, Lever uses hex IDs
- `last_seen` timestamp is updated to now for postings still on the board; postings that disappear keep their old `last_seen`, which drifts behind — that's how staleness is detected
- Only postings that pass the keyword pre-filter (Story 1) are stored; staleness tracking uses all fetched IDs from the ATS regardless of filter
