# Story 6: Company List Management

**As a job seeker, I want to add or remove companies from my watch list**, managed through the database.

## Acceptance Criteria

- [ ] `job-monitor import <csv>` seeds the companies table from an existing CSV (one-time migration)
- [ ] `job-monitor add <company>` adds a company and auto-detects its ATS
- [ ] `job-monitor remove <company>` marks a company as inactive (soft delete)
- [ ] `job-monitor list` shows all active companies with their ATS status
- [ ] New companies with unknown ATS can be auto-detected using existing `ats-detect` probing logic

## Notes

- Reuse slug generation and ATS probing from `ats-detect`
- The `companies` table in SQLite is the single source of truth after initial import
