---
title: Pipeline behavioral tests for the run command
status: todo
story: "2"
---

The current test suite covers individual store operations and CLI subcommands but nothing exercises the pipeline flow end-to-end. Bugs like re-scoring all postings on every run, re-classifying already-classified titles, and not clearing notified_at on rescore were all missed because the pieces work individually but compose incorrectly.

## Tasks

- [ ] Add a fake/stub LLM client interface so `FilterTitles` and `ScoreContent` can be called without hitting the API
- [ ] Add `TestRunPipeline` in `cmd_test.go` covering the default flow: fetch → classify only unclassified → score only unscored → output matches → mark notified
- [ ] Add `TestRunPipelineReclassify` covering the reclassify flow: previously scored+notified postings get rescored and appear in matches again
- [ ] Verify that a second run with no new postings produces no LLM calls (no reclassification, no rescoring)
