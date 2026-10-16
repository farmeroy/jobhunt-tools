---
title: Extract shared ATS utilities into internal/atsutil
status: done
story: "1, 6"
---

Extract `slugify()`, `checkProvider()`, `detectATS()`, and the HTTP client from `ats-detect/main.go` into `internal/atsutil/` so both `ats-detect` and `job-monitor` can share them. Refactor `ats-detect` to import the shared package.

## Done

- `internal/atsutil/slugify.go` — exported `Slugify()`
- `internal/atsutil/probe.go` — exported `Provider`, `CheckProvider()`, `DetectATS()`, `NewHTTPClient()`
- `internal/atsutil/slugify_test.go` — 7 unit tests
- `internal/atsutil/probe_test.go` — 7 tests using `httptest` (no network)
- `ats-detect/main.go` refactored to import `internal/atsutil`
