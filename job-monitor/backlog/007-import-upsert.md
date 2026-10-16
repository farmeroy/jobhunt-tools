---
title: Import upsert / update behavior
status: todo
story: "6"
---

Handle reimport conflicts gracefully instead of just skipping duplicates.

## Tasks

- [ ] Decide UX: `--upsert` flag on import, or separate `update` command
- [ ] Warn user before overwriting existing data (ATS, slug, reactivate?)
- [ ] Implement chosen approach with tests
