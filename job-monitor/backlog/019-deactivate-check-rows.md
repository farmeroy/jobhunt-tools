---
title: DeactivateCompany should verify the company exists
status: done
story: "6"
---

`DeactivateCompany` runs an UPDATE that silently matches zero rows on a typo. The CLI prints success regardless.

## Tasks

- [ ] Check `RowsAffected` after the UPDATE
- [ ] Return an error (or a "not found" sentinel) when zero rows matched
- [ ] Surface the error in the `remove` command
