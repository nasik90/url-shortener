package pg

import (
	"database/sql"
)

type Store struct {
	conn *sql.DB
}

func NewStore(conn *sql.DB) *Store {
	return &Store{conn: conn}
}

func (s *Store) SaveShortURL(shortURL, originalURL string) error {

	return nil
}

func (s *Store) GetOriginalURL(shortURL string) (string, error) {

	return "", nil
}

func (s *Store) IsUnique(shortURL string) bool {
	return true
}

func (s *Store) Ping() error {
	return s.conn.Ping()
}
