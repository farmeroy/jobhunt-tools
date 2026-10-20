---
title: Title filter prompt tuning — use configured target roles
status: todo
story: "3"
depends: "006"
---

The current two-step title filter infers target roles from the resume, but gets seniority wrong (passes Staff/Principal titles). The title filter should be a **wide net** — the content filter (Story 4) is the strict pass.

Once backlog 006 lands, the title classifier can use explicit target roles from config instead of (or alongside) inferring them from the resume.

## Tasks

- [ ] Use `target_roles` from config in the classification prompt
- [ ] Loosen classification prompt: err on the side of including rather than excluding
- [ ] Evaluate whether the two-step prompt is still needed when target roles are configured
- [ ] Build eval set from real classification mistakes (log files have the data)
- [ ] Investigate seniority mismatch problem: "Senior Backend Engineer" passes both title filter and content scorer. Target roles config may help — if user specifies "Mid-level Backend Engineer", the LLM has a clearer signal to deprioritize Senior roles
