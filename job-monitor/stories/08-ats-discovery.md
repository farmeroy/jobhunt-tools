# Story 8 (Stretch): ATS Discovery

**As a job seeker, I want the tool to discover career page URLs for companies without a known ATS**, expanding coverage beyond the initial ~77.

## Acceptance Criteria

- [ ] `job-monitor discover` attempts to find career pages for companies with no ATS
- [ ] Uses web search or common URL patterns (e.g., `{company}.com/careers`) to locate pages
- [ ] Discovered ATS or career URLs are saved to the companies table
- [ ] Results can be reviewed and confirmed before activating monitoring

## Notes

- The existing `ats-detect` only probes Greenhouse, Ashby, Lever — this extends to other providers and direct career pages
- Could also probe for Workday, iCIMS, JazzHR patterns
- Low priority — the ~77 API-friendly companies are the high-value target
