---
title: Redesign content scoring — structured output, better prompt, model upgrade
status: todo
story: "5"
---

Content scoring produces verbose, essay-like reasoning that isn't useful for evaluating score calibration. The prompt's recruiter persona encourages hedging — scores of 7-9 with paragraphs of caveats that contradict the score (e.g. 9/10 for a role requiring 8+ years when the candidate has 3). Reasoning exists for eval, so it needs to be structured and measurable.

## Tasks

- [ ] Replace `reasoning` string with structured `met`, `unmet`, `dealbreakers` arrays in tool schema
- [ ] Update `ContentResult` type to carry the three lists
- [ ] Join lists into a compact format for DB storage (no migration — `content_reasoning` column stays TEXT)
- [ ] Rewrite system prompt with calibrated scoring guide and decision rules (dealbreakers force ≤ 3, nice-to-haves aren't dealbreakers, experience gap > 2 years is a dealbreaker)
- [ ] Upgrade model from Haiku 4.5 → Haiku 5.5 ($0.10/$0.50 vs $1.00/$5.00 per MTok, better quality)
