---
title: Redesign content scoring — structured output, better prompt, model upgrade
status: partial
story: "4"
---

Content scoring produced verbose, essay-like reasoning that wasn't useful for evaluating score calibration. The prompt's recruiter persona encouraged hedging — scores of 7-9 with paragraphs of caveats that contradicted the score.

## Done

- [x] Replace `reasoning` string with structured `met`, `unmet`, `dealbreakers` arrays in tool schema
- [x] Update `ContentResult` type to carry the three lists
- [x] Join lists into a compact format for DB storage (`+ met | - unmet | ! dealbreaker`)
- [x] Upgrade model to Sonnet 5.5 (better quality for nuanced scoring)

## Remaining

- [ ] Rewrite system prompt with calibrated scoring guide and decision rules (dealbreakers force ≤ 3, nice-to-haves aren't dealbreakers, experience gap > 2 years is a dealbreaker)
