# Story 2: Dedup & New Post Detection

**As a job seeker, I want the tool to track which postings it has already seen**, so I only get notified about genuinely new opportunities.

## Acceptance Criteria

- [ ] SQLite database stores all fetched postings with a unique key (company + ATS job ID)
- [ ] On each fetch cycle, only postings not already in the database are flagged as "new"
- [ ] Previously seen postings are updated if their title or description changed
- [ ] Postings that disappear from a company's board are marked as closed/removed (not deleted)
- [ ] `first_seen` timestamp recorded for every posting

## Notes

- The dedup key needs to be ATS-specific — Greenhouse uses numeric IDs, Ashby uses UUIDs, Lever uses hex IDs
- Consider a `last_seen` timestamp to detect removed postings
