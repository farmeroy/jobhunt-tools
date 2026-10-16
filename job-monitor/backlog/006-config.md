---
title: Config file loading
status: todo
story: "1"
---

YAML config file with paths, API keys, and settings.

## Tasks

- [ ] Config struct — db_path, resume_path, anthropic_api_key, concurrency, score_threshold, notification settings
- [ ] YAML loader with env var fallback for API key (`ANTHROPIC_API_KEY`)
- [ ] `config.example.yaml` with documented defaults
