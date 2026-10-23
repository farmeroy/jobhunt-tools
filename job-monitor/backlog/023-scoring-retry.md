---
title: Retry content scoring on malformed LLM responses
status: done
story: "2"
---

When the LLM returns a malformed tool response (e.g. a string instead of `[]string` for the `met` field), `scorePosting` returns a parse error. The caller in `ScoreContent` logs it and moves on — no retry. The SDK's `WithMaxRetries(5)` only covers transient HTTP errors, not application-level failures.

A second attempt usually gets valid output, so these postings are silently dropped for no good reason.

## Tasks

- [x] Add retry loop around `scorePosting` for non-API errors (e.g. JSON parse failures)
- [x] Cap retries (3 attempts) to avoid infinite loops on consistently bad responses

## Commits

- 187a903 fix: retry content scoring on malformed LLM responses

## Notes

- At 112 concurrent scoring calls with 10-wide parallelism, we're hitting ~88% of the Haiku rate limit. Retries add marginal load but could push over during large initial loads. Not a concern for steady-state runs.
