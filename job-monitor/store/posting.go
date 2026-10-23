package store

import (
	"strings"
	"time"
)

func (s *Store) UpsertPostings(companyID int64, postings []Posting) (int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var countBefore int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM postings WHERE company_id = ?`, companyID).Scan(&countBefore); err != nil {
		return 0, err
	}

	now := time.Now().UTC()
	for _, p := range postings {
		_, err := tx.Exec(
			`INSERT INTO postings (company_id, external_id, title, url, location, description, first_seen, last_seen)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			 ON CONFLICT(company_id, external_id) DO UPDATE SET
			   title = excluded.title, url = excluded.url, location = excluded.location,
			   description = excluded.description, last_seen = excluded.last_seen`,
			companyID, p.ExternalID, p.Title, p.URL, p.Location, p.Description, now, now,
		)
		if err != nil {
			return 0, err
		}
	}

	var countAfter int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM postings WHERE company_id = ?`, companyID).Scan(&countAfter); err != nil {
		return 0, err
	}

	return countAfter - countBefore, tx.Commit()
}

func (s *Store) RefreshLastSeen(companyID int64, currentExternalIDs []string) (int64, error) {
	if len(currentExternalIDs) == 0 {
		return 0, nil
	}

	placeholders := strings.Repeat("?,", len(currentExternalIDs))
	placeholders = placeholders[:len(placeholders)-1]

	now := time.Now().UTC()
	args := []any{now, companyID}
	for _, id := range currentExternalIDs {
		args = append(args, id)
	}

	query := `UPDATE postings SET last_seen = ?
		WHERE company_id = ? AND external_id IN (` + placeholders + `)`
	res, err := s.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

type PostingOrder int

const (
	OrderByFirstSeen PostingOrder = iota
	OrderByScoreDesc
)

type PostingFilter struct {
	Company           string
	TitleUnclassified bool
	TitleRelevant     *bool
	Unscored          bool
	ScoredSince       *time.Time
	ContentScoreMin   *float64
	Notified          *bool
	OrderBy           PostingOrder
}

func (s *Store) ListPostings(f PostingFilter) ([]Posting, error) {
	query := `SELECT p.id, p.company_id, c.name, p.external_id, p.title, p.url, p.location, p.description,
		p.first_seen, p.last_seen, p.title_relevant, p.content_score, p.content_reasoning,
		p.scored_on, p.notified_at, p.applied_at
	 FROM postings p
	 JOIN companies c ON c.id = p.company_id
	 WHERE 1=1`
	var args []any

	if f.Company != "" {
		query += ` AND c.name LIKE ?`
		args = append(args, "%"+f.Company+"%")
	}
	if f.TitleUnclassified {
		query += ` AND p.title_relevant IS NULL`
	} else if f.TitleRelevant != nil {
		query += ` AND p.title_relevant = ?`
		args = append(args, *f.TitleRelevant)
	}
	if f.ScoredSince != nil {
		query += ` AND (p.scored_on IS NULL OR p.scored_on >= ?)`
		args = append(args, *f.ScoredSince)
	} else if f.Unscored {
		query += ` AND p.scored_on IS NULL`
	}
	if f.ContentScoreMin != nil {
		query += ` AND p.content_score >= ?`
		args = append(args, *f.ContentScoreMin)
	}
	if f.Notified != nil {
		if *f.Notified {
			query += ` AND p.notified_at IS NOT NULL`
		} else {
			query += ` AND p.notified_at IS NULL`
		}
	}

	switch f.OrderBy {
	case OrderByScoreDesc:
		query += ` ORDER BY p.content_score DESC`
	default:
		query += ` ORDER BY p.first_seen DESC`
	}

	return s.queryPostings(query, args...)
}

func (s *Store) GetDistinctTitles(since time.Time) ([]string, error) {
	query := `SELECT DISTINCT title FROM postings`
	var args []any

	if !since.IsZero() {
		query += ` WHERE first_seen >= ?`
		args = append(args, since)
	} else {
		query += ` WHERE title_relevant IS NULL`
	}

	query += ` ORDER BY title`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var titles []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		titles = append(titles, t)
	}
	return titles, rows.Err()
}

func (s *Store) SetTitleRelevanceByTitle(title string, relevant bool) (int64, error) {
	res, err := s.db.Exec(`UPDATE postings SET title_relevant = ? WHERE title = ?`, relevant, title)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (s *Store) SetTitleRelevance(postingID int64, relevant bool) error {
	_, err := s.db.Exec(`UPDATE postings SET title_relevant = ? WHERE id = ?`, relevant, postingID)
	return err
}

func (s *Store) SetContentScore(postingID int64, score float64, reasoning string, clearNotified bool) error {
	now := time.Now().UTC()
	if clearNotified {
		_, err := s.db.Exec(
			`UPDATE postings SET content_score = ?, content_reasoning = ?, scored_on = ?, notified_at = NULL WHERE id = ?`,
			score, reasoning, now, postingID,
		)
		return err
	}
	_, err := s.db.Exec(
		`UPDATE postings SET content_score = ?, content_reasoning = ?, scored_on = ? WHERE id = ?`,
		score, reasoning, now, postingID,
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
			&p.ID, &p.CompanyID, &p.CompanyName, &p.ExternalID, &p.Title, &p.URL, &p.Location, &p.Description,
			&p.FirstSeen, &p.LastSeen, &p.TitleRelevant, &p.ContentScore, &p.ContentReasoning,
			&p.ScoredOn, &p.NotifiedAt, &p.AppliedAt,
		); err != nil {
			return nil, err
		}
		postings = append(postings, p)
	}
	return postings, rows.Err()
}
