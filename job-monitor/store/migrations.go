package store

const schemaSQL = `
CREATE TABLE IF NOT EXISTS companies (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	slug TEXT NOT NULL DEFAULT '',
	ats TEXT,
	active INTEGER NOT NULL DEFAULT 1,
	added_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS postings (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	company_id INTEGER NOT NULL REFERENCES companies(id),
	external_id TEXT NOT NULL,
	title TEXT NOT NULL,
	url TEXT,
	description TEXT,
	first_seen DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	last_seen DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	title_relevant INTEGER,
	content_score REAL,
	content_reasoning TEXT,
	notified_at DATETIME,
	applied_at DATETIME,
	UNIQUE(company_id, external_id)
);

CREATE INDEX IF NOT EXISTS idx_postings_company ON postings(company_id);
CREATE INDEX IF NOT EXISTS idx_postings_filter ON postings(title_relevant, content_score, notified_at);
`
