package store

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

type Store struct {
	db *sql.DB
}

func Open(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	provider, err := goose.NewProvider(goose.DialectSQLite3, s.db, migrations)
	if err != nil {
		return err
	}
	_, err = provider.Up(context.Background())
	return err
}
