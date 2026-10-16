# Story 5: Notification

**As a job seeker, I want to be notified of matching jobs with company name, title, match reasoning, and application link**, so I can apply quickly.

## Acceptance Criteria

- [ ] Matched postings are sent as an email digest
- [ ] Email includes: company name, job title, match score, reasoning summary, and direct apply link
- [ ] Postings are grouped by company
- [ ] Each posting is only notified once (tracked via `notified` flag in DB)
- [ ] If no new matches are found, no email is sent
- [ ] Notification method is behind an interface for future extensibility (Slack, Discord, etc.)

## Notes

- Email sending options: SMTP directly, or a service like SendGrid/Mailgun for reliability
- Consider a simple HTML template for readability
