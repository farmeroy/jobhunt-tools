package store

func (s *Store) AddCompany(name, slug, ats string) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO companies (name, slug, ats) VALUES (?, ?, ?)`,
		name, slug, ats,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) DeactivateCompany(name string) error {
	_, err := s.db.Exec(`UPDATE companies SET active = 0 WHERE name = ?`, name)
	return err
}

func (s *Store) ListActiveCompanies() ([]Company, error) {
	return s.queryCompanies(`SELECT id, name, slug, ats, active, added_at FROM companies WHERE active = 1 ORDER BY name`)
}

func (s *Store) ListAllCompanies() ([]Company, error) {
	return s.queryCompanies(`SELECT id, name, slug, ats, active, added_at FROM companies ORDER BY name`)
}

func (s *Store) UpdateSlug(companyID int64, slug string) error {
	_, err := s.db.Exec(`UPDATE companies SET slug = ? WHERE id = ?`, slug, companyID)
	return err
}

func (s *Store) queryCompanies(query string, args ...any) ([]Company, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companies []Company
	for rows.Next() {
		var c Company
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.ATS, &c.Active, &c.AddedAt); err != nil {
			return nil, err
		}
		companies = append(companies, c)
	}
	return companies, rows.Err()
}
