# Job Monitor — Architecture

## Overview

A CLI tool that monitors target companies for new job postings, filters them for relevance using an LLM, and notifies the user of matches.

### Pipeline

```
[Scheduled Trigger]
       ↓
[Job Fetcher]  — hit ATS APIs for each company
       ↓
[Keyword Pre-Filter]  — title whitelist (engineer, developer, software, ...) + location whitelist (remote, US)
       ↓
[Dedup]  — compare against previously seen postings, surface only new ones
       ↓
[LLM Filter — Pass 1: Title]  — cheap/fast: is this a relevant role?
       ↓
[LLM Filter — Pass 2: Content]  — for title-pass jobs: score fit against CV/preferences
       ↓
[Notification]  — alert with company, title, match reasoning, and apply link
```

## Tech Stack

| Component | Choice | Rationale |
|-----------|--------|-----------|
| Language | Go | Existing Go tooling in repo (`ats-detect`), good for concurrent HTTP + scheduled work, building Go skills |
| Storage | SQLite | Local, single-user, queryable, no external deps. Stores seen postings, filter results, logs |
| LLM | Haiku 4.5 (Anthropic API) | ~$0.80/MTok input, fast, sufficient for title classification and description scoring |
| Notification | Email digest | Reliable, works everywhere, includes links. Daily summary grouped by company |
| Scheduling | Cron (system) | Simple, standard. CLI also supports manual one-off runs |

## ATS Coverage

### Phase 1 — Public JSON APIs (~77 companies)

| ATS | ~Count | API Pattern |
|-----|--------|-------------|
| Greenhouse | 40 | `boards-api.greenhouse.io/v1/boards/{slug}/jobs` |
| Ashby | 30 | `jobs.ashbyhq.com/api/non-user-graphql` |
| Lever | 7 | `api.lever.co/v0/postings/{slug}` |

### Phase 2+ — Harder ATS (~21 companies)

Workday, iCIMS, JazzHR, Rippling, Workable, Taleo — mix of semi-public APIs and scraping.

### Unresolved — No ATS identified (~80 companies)

Need career page discovery. Could be manual curation or automated search.

## LLM Filter Design

### Pass 1: Title Filter
- Batch many job titles into a single API call
- Binary classification: relevant (SWE, backend, fullstack, etc.) vs. irrelevant (marketing, sales, design, etc.)
- Very cheap — one call per company fetch cycle

### Pass 2: Content Filter
- Runs only on jobs that pass the title filter
- Scores job description against user's CV and stated preferences
- Returns: match score, reasoning, key requirements met/unmet
- One call per job posting (still cheap — only a fraction of total postings reach this stage)

## Data Model (SQLite)

SQLite is the single source of truth for both companies and postings. The existing `companies.csv` is imported once as a seed; after that, companies are managed through the CLI/database.

### `companies` table
- `id` — auto-increment
- `name` — company name
- `slug` — URL-friendly name (for ATS API calls)
- `ats` — ATS provider (greenhouse, ashby, lever, etc. or null if unknown)
- `active` — boolean (whether to monitor this company)
- `added_at` — timestamp

### `postings` table
- `id` — auto-increment
- `company_id` — FK to companies
- `external_id` — ATS-specific job ID
- `title` — job title
- `url` — application link
- `description` — full job description (if fetched)
- `first_seen` — timestamp
- `last_seen` — timestamp (to detect removed postings)
- `title_relevant` — boolean (pass 1 result, nullable)
- `content_score` — float (pass 2 result, nullable)
- `content_reasoning` — text (pass 2 reasoning, nullable)
- `notified_at` — timestamp (nullable, when the user was notified)
- `applied_at` — timestamp (nullable, when the user marked as applied)

Unique constraint on `(company_id, external_id)` for dedup.

## LLM Context

The user's resume is loaded once and included in the system prompt for both filter passes. No separate relevance config — the resume *is* the relevance context.

### Keyword Pre-Filter (at fetch time)

Before postings reach the database, a cheap keyword filter runs:
- **Title whitelist**: posting title must contain one of: engineer, developer, software, sre, devops, platform, infrastructure, backend, frontend, fullstack, full-stack (case-insensitive)
- **Location whitelist**: location must contain one of: remote, united states, US, california, san francisco, new york, hybrid — or be empty/missing (pass by default)

This dramatically reduces volume before any LLM calls. The company list is curated for remote-friendly companies, so US-based postings are assumed remote-eligible. Lists are hardcoded initially, configurable later via config file.

## Configuration

Minimal config file (YAML or TOML):
- Path to resume (PDF or text)
- Notification settings: email address
- Anthropic API key (or env var)
- SQLite database path

## Project Structure

```
job-monitor/
├── main.go
├── go.mod
├── config.yaml
├── fetcher/        # ATS-specific fetchers (greenhouse.go, ashby.go, lever.go)
├── store/          # SQLite persistence layer
├── filter/         # LLM filtering logic (title + content passes)
├── notify/         # Notification (email first, extensible)
└── config/         # Config loading and validation
```

## Seed Data

The existing `companies.csv` is imported once via `job-monitor import` to populate the `companies` table. After that, the database is the source of truth and companies are managed via CLI (`job-monitor add`, `job-monitor remove`, `job-monitor list`).
