---
title: E2E fetcher tests against live APIs
status: todo
story: "1"
---

Add real API tests gated behind a build tag so they don't run in CI.

## Tasks

- [ ] `//go:build e2e` test file hitting live Greenhouse, Lever, Ashby APIs with known slugs
- [ ] Verify response parsing against actual JSON structure
- [ ] Run manually via `go test -tags e2e ./job-monitor/fetcher/...`
