---
title: Config file loading
status: deferred
story: "1"
---

YAML config file with paths, API keys, and settings. Currently deferred — flags and env vars are sufficient.

## Current approach

- `--db` flag for database path
- `--resume` flag for resume file path
- `ANTHROPIC_API_KEY` env var for API key

## If needed later

- [ ] Config struct — db_path, resume_path, concurrency, score_threshold, notification settings
- [ ] YAML loader with env var fallback for API key
- [ ] `config.example.yaml` with documented defaults
- [ ] API key should NOT be in the config file if it's in version control
