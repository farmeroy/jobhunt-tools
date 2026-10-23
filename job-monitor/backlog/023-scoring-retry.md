---
title: Retry content scoring on malformed LLM responses
status: todo
story: "2"
---

When the LLM returns a malformed tool response (e.g. a string instead of `[]string` for the `met` field), `scorePosting` returns a parse error. The caller in `ScoreContent` logs it and moves on — no retry. The SDK's `WithMaxRetries(5)` only covers transient HTTP errors, not application-level failures.

A second attempt usually gets valid output, so these postings are silently dropped for no good reason.

## Tasks

- [ ] Add retry loop around `scorePosting` for non-API errors (e.g. JSON parse failures)
- [ ] Cap retries (2-3 attempts) to avoid infinite loops on consistently bad responses
