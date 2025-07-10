package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"url-shortener/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "sqlite.sqlite.New"

	db, err := sql.Open("sqlite", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to open database: %w", op, err)
	}

	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS url (
 	   	id INTEGER PRIMARY KEY,
    	alias TEXT NOT NULL UNIQUE,
    	url TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to prepare statement: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to execute statement: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "sqlite.Storage.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO url(alias, url) VALUES (?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: failed to prepare statement: %w", op, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(alias, urlToSave)
	if err != nil {
		if errors.Is(err, storage.ErrURLExists) || err.Error() == "UNIQUE constraint failed: url.alias" {
			return 0, storage.ErrURLExists
		}

		return 0, fmt.Errorf("%s: failed to execute statement: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "sqlite.Storage.GetURL"

	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s: failed to prepare statement: %w", op, err)
	}
	defer stmt.Close()

	var resURL string
	err = stmt.QueryRow(alias).Scan(&resURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s: failed to execute query: %w", op, err)
	}

	return resURL, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "sqlite.Storage.DeleteURL"

	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias = ?")
	if err != nil {
		return fmt.Errorf("%s: failed to prepare statement: %w", op, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(alias)
	if err != nil {
		return fmt.Errorf("%s: failed to execute statement: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return storage.ErrURLNotFound
	}

	return nil
}
