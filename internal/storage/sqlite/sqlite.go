package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"url-shortner/internal/storage"

	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const fn = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS url(
		id INTEGER PRIMARY KEY,
		alias TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const fn = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO url(url, alias) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	res, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", fn, storage.ErrURLExists)
		}
		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get insert id: %w", fn, err)
	}

	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const fn = "storage.sqlite.GetURL"

	stmt, err := s.db.Prepare("SELECT url from url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s: %w", fn, err)
	}

	var urlToGet string
	err = stmt.QueryRow(alias).Scan(&urlToGet)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s: %w", fn, err)
	}

	return urlToGet, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const fn = "storage.sqlite.DeleteURL"

	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	_, err = stmt.Exec(alias)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}
