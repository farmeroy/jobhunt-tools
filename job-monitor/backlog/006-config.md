---
title: Config file for target roles and allowed locations
status: todo
story: "1, 3"
---

YAML config file with target roles and location allowlist. Flags and env vars remain for paths and API keys.

## Motivation

- The title classifier infers target roles from the resume, but gets seniority wrong (e.g. passes Staff titles). Explicit target roles give it better signal.
- The location pre-filter uses a hardcoded keyword list that lets through Canadian "Remote" postings. A config-driven allowlist makes this tunable.

## Tasks

- [ ] Config struct with `target_roles` ([]string) and `allowed_locations` ([]string)
- [ ] YAML loader — default path `config.yaml` next to the database, overridable with `--config` flag
- [ ] Feed `target_roles` into the title classifier prompt (replaces or supplements the "infer from resume" step)
- [ ] Feed `allowed_locations` into content scorer (currently hardcoded as `["remote US", "hybrid US"]` in `cmd/run.go`)
- [ ] `config.example.yaml` with documented defaults

Note: `allowed_locations` is a semantic allowlist for the LLM content scorer, not the keyword pre-filter. The LLM uses geographic reasoning (e.g. "Remote California" matches "Remote US").
