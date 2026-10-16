package store

import (
	"database/sql"
	"strings"
	"time"
)

func (s *Store) UpsertPostings(companyID int64, postings []Posting) (int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	newCount := 0
	now := time.Now().UTC()

	for _, p := range postings {
		var existingID int64
		err := tx.QueryRow(
			`SELECT id FROM postings WHERE company_id = ? AND external_id = ?`,
			companyID, p.ExternalID,
		).Scan(&existingID)

		if err == sql.ErrNoRows {
			_, err = tx.Exec(
				`INSERT INTO postings (company_id, external_id, title, url, description, first_seen, last_seen)
				 VALUES (?, ?, ?, ?, ?, ?, ?)`,
				companyID, p.ExternalID, p.Title, p.URL, p.Description, now, now,
			)
			if err != nil {
				return 0, err
			}
			newCount++
		} else if err != nil {
			return 0, err
		} else {
			_, err = tx.Exec(
				`UPDATE postings SET title = ?, url = ?, description = ?, last_seen = ? WHERE id = ?`,
				p.Title, p.URL, p.Description, now, existingID,
			)
			if err != nil {
				return 0, err
			}
		}
	}

	return newCount, tx.Commit()
}

func (s *Store) MarkStale(companyID int64, currentExternalIDs []string) error {
	if len(currentExternalIDs) == 0 {
		// Board is empty — all postings for this company are stale
		_, err := s.db.Exec(
			`UPDATE postings SET last_seen = last_seen WHERE company_id = ?`,
			companyID,
		)
		return err
	}

	placeholders := strings.Repeat("?,", len(currentExternalIDs))
	placeholders = placeholders[:len(placeholders)-1]

	args := []any{companyID}
	for _, id := range currentExternalIDs {
		args = append(args, id)
	}

	query := `UPDATE postings SET last_seen = last_seen
		WHERE company_id = ? AND external_id NOT IN (` + placeholders + `)`
	_, err := s.db.Exec(query, args...)
	return err
}

func (s *Store) GetNewUnfilteredPostings() ([]Posting, error) {
	return s.queryPostings(
		`SELECT p.id, p.company_id, c.name, p.external_id, p.title, p.url, p.description,
			p.first_seen, p.last_seen, p.title_relevant, p.content_score, p.content_reasoning,
			p.notified_at, p.applied_at
		 FROM postings p
		 JOIN companies c ON c.id = p.company_id
		 WHERE p.title_relevant IS NULL
		 ORDER BY p.first_seen DESC`,
	)
}

func (s *Store) GetTitlePassedUnscored() ([]Posting, error) {
	return s.queryPostings(
		`SELECT p.id, p.company_id, c.name, p.external_id, p.title, p.url, p.description,
			p.first_seen, p.last_seen, p.title_relevant, p.content_score, p.content_reasoning,
			p.notified_at, p.applied_at
		 FROM postings p
		 JOIN companies c ON c.id = p.company_id
		 WHERE p.title_relevant = 1 AND p.content_score IS NULL
		 ORDER BY p.first_seen DESC`,
	)
}

func (s *Store) GetUnnotifiedMatches(threshold float64) ([]Posting, error) {
	return s.queryPostings(
		`SELECT p.id, p.company_id, c.name, p.external_id, p.title, p.url, p.description,
			p.first_seen, p.last_seen, p.title_relevant, p.content_score, p.content_reasoning,
			p.notified_at, p.applied_at
		 FROM postings p
		 JOIN companies c ON c.id = p.company_id
		 WHERE p.content_score >= ? AND p.notified_at IS NULL
		 ORDER BY p.content_score DESC`,
		threshold,
	)
}

func (s *Store) SetTitleRelevance(postingID int64, relevant bool) error {
	_, err := s.db.Exec(`UPDATE postings SET title_relevant = ? WHERE id = ?`, relevant, postingID)
	return err
}

func (s *Store) SetContentScore(postingID int64, score float64, reasoning string) error {
	_, err := s.db.Exec(
		`UPDATE postings SET content_score = ?, content_reasoning = ? WHERE id = ?`,
		score, reasoning, postingID,
	)
	return err
}

func (s *Store) MarkNotified(postingIDs []int64) error {
	if len(postingIDs) == 0 {
		return nil
	}
	placeholders := strings.Repeat("?,", len(postingIDs))
	placeholders = placeholders[:len(placeholders)-1]

	args := []any{time.Now().UTC()}
	for _, id := range postingIDs {
		args = append(args, id)
	}

	_, err := s.db.Exec(
		`UPDATE postings SET notified_at = ? WHERE id IN (`+placeholders+`)`,
		args...,
	)
	return err
}

func (s *Store) queryPostings(query string, args ...any) ([]Posting, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var postings []Posting
	for rows.Next() {
		var p Posting
		if err := rows.Scan(
			&p.ID, &p.CompanyID, &p.CompanyName, &p.ExternalID, &p.Title, &p.URL, &p.Description,
			&p.FirstSeen, &p.LastSeen, &p.TitleRelevant, &p.ContentScore, &p.ContentReasoning,
			&p.NotifiedAt, &p.AppliedAt,
		); err != nil {
			return nil, err
		}
		postings = append(postings, p)
	}
	return postings, rows.Err()
}
