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
| LLM (title) | Haiku 4.5 (Anthropic API) | ~$0.80/MTok input, fast, sufficient for title classification |
| LLM (content) | Sonnet 5.5 (Anthropic API) | Better quality for nuanced job description scoring |
| Notification | Text file (temporary) → Email digest (planned) | Matches written to `matches.txt` grouped by company. Email notifications planned |
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
- **Distinct titles** — classifies unique title strings, not individual postings. Results are bulk-applied to all postings with that title
- **Two-step prompt** — (1) infer target roles from resume, (2) classify titles against those targets. Keeps classification focused and cheaper
- Batch titles into a single API call (chunked at 100 per call), numeric indices for robust round-tripping
- Binary classification: relevant vs. irrelevant, using Anthropic tool_use for structured output
- `--reclassify <duration>` re-runs classification on postings within a time period (e.g. `7d`, `2w`)
- Default run classifies all distinct titles; reclassify mode filters by `first_seen`
- Token usage tracked per run (input tokens, output tokens, API call count)

### Pass 2: Content Filter
- Runs only on jobs that pass the title filter
- Scores job description against user's CV and stated preferences using Sonnet 5.5
- Returns: match score (1-10), structured `met`/`unmet`/`dealbreakers` arrays via tool_use
- Reasoning stored as compact text format (`+ met | - unmet | ! dealbreaker`)
- `scored_on` timestamp enables rescoring via `--reclassify` (same window as title reclassification)
- One call per job posting (still cheap — only a fraction of total postings reach this stage)

## Schema Migrations

Schema is managed with [goose](https://github.com/pressly/goose) (`pressly/goose/v3`). SQL migration files live in `store/*.sql` and are embedded into the binary via `//go:embed`. Migrations run automatically on `store.Open()` using goose's `NewProvider` API. To add a new migration, create `store/NNN_description.sql` with `-- +goose Up` / `-- +goose Down` annotations.

## Data Model (SQLite)

SQLite is the single source of truth for both companies and postings. The existing `companies.csv` is imported once as a seed; after that, companies are managed through the CLI/database.

### `companies` table
- `id` — auto-increment
- `name` — company name (unique)
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
- `location` — job location (from ATS, used for pre-filtering)
- `description` — full job description (if fetched)
- `first_seen` — timestamp
- `last_seen` — timestamp (to detect removed postings)
- `title_relevant` — boolean (pass 1 result, nullable)
- `content_score` — float (pass 2 result, nullable)
- `content_reasoning` — text (pass 2 reasoning, nullable)
- `scored_on` — timestamp (nullable, when content was last scored)
- `notified_at` — timestamp (nullable, when the user was notified)
- `applied_at` — timestamp (nullable, when the user marked as applied)

Unique constraint on `(company_id, external_id)` for dedup.

## LLM Context

The user's resume is loaded once and included in the system prompt for both filter passes. No separate relevance config — the resume *is* the relevance context.

### Resume Loading

Resume is loaded via `resume.LoadText(path)`. Supports `.pdf`, `.txt`, and `.md`. PDF text extraction uses `pdftotext` (poppler-utils). The extracted text is cached in memory for the duration of the run — no re-parsing per call.

### Keyword Pre-Filter (at fetch time)

Before postings reach the database, a cheap keyword filter runs:
- **Title whitelist**: posting title must contain one of: engineer, developer, software, sre, devops, platform, infrastructure, backend, frontend, fullstack, full-stack (case-insensitive)
- **Location whitelist**: location must contain one of: remote, united states, US, california, san francisco, new york, hybrid — or be empty/missing (pass by default)

This dramatically reduces volume before any LLM calls. The company list is curated for remote-friendly companies, so US-based postings are assumed remote-eligible. Lists are hardcoded initially, configurable later via config file.

## Configuration

- `--db` flag — SQLite database path (default: `job-monitor.db`)
- `--resume` flag — path to resume file (PDF, txt, or md)
- `ANTHROPIC_API_KEY` env var — required for LLM filter commands

### Design Decisions

- **`run` is the full pipeline** — fetch → classify → (eventually) notify. Designed for unattended cron execution.
- **Classification runs on all distinct titles** by default. `--reclassify` re-runs within a time window.
- **No config file** — flags and env vars are sufficient for now. API key from env, resume path from flag.
- **Run logs** — each `run` writes a timestamped log file to `.logs/` alongside the DB. All classification results, token usage, and errors are persisted.

## Project Structure

```
job-monitor/
├── main.go
├── go.mod
├── cmd/            # Cobra CLI commands (run, import, add, remove, list, postings)
├── fetcher/        # ATS-specific fetchers (greenhouse.go, ashby.go, lever.go) + keyword pre-filter
├── filter/         # LLM filtering logic (title + content passes)
├── resume/         # Resume text extraction (PDF via pdftotext, txt, md)
├── store/          # SQLite persistence layer
└── notify/         # Match output (text file for now, email planned)
```

## Seed Data

The existing `companies.csv` is imported once via `job-monitor import` to populate the `companies` table. After that, the database is the source of truth and companies are managed via CLI (`job-monitor add`, `job-monitor remove`, `job-monitor list`).
