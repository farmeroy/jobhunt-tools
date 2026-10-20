---
title: Title filter prompt tuning — too strict
status: todo
story: "3"
---

The current two-step title filter (infer targets → classify) is too strict. It rejects plausibly relevant titles like "Analytics Engineer II, Full Stack" and "Engineering Manager, AI Applications".

The title filter should be a **wide net** — catch anything plausibly relevant. The content filter (Story 4) is where strict matching belongs.

## Tasks

- [ ] Loosen classification prompt: err on the side of including rather than excluding
- [ ] Consider whether the two-step prompt is worth it, or if a single prompt with a broader instruction works better
- [ ] Build eval set from real classification mistakes (log files have the data)
