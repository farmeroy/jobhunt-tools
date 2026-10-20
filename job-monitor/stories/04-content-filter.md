# Story 4: LLM Content Filter (Pass 2)

**As a job seeker, I want title-passing postings scored against my CV and preferences**, so I can prioritize the best-fit roles.

## Acceptance Criteria

- [ ] Postings that pass the title filter have their full description fetched (if not already)
- [x] Each description is scored against the user's CV and stated preferences using Haiku 4.5
- [x] Returns a match score (e.g., 1-10) and brief reasoning
- [x] Reasoning highlights key requirements met and unmet
- [x] Score and reasoning are stored in the database
- [ ] Only postings above a configurable score threshold proceed to notification

## Notes

- Resume is loaded once and included in each prompt as context (same as title filter)
- The LLM judges fit holistically from the resume — no separate config of preferences needed
- One API call per posting — but only title-relevant postings reach this stage
