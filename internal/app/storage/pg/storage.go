package pg

import (
	"context"
	"database/sql"
)

type Store struct {
	conn *sql.DB
}

func NewStore(conn *sql.DB) (*Store, error) {
	s := &Store{conn: conn}
	err := s.Bootstrap(context.Background())
	if err != nil {
		return s, err
	}
	return s, nil
}

func (s Store) Close() error {
	return s.conn.Close()
}

// Bootstrap подготавливает БД к работе, создавая необходимые таблицы и индексы
func (s Store) Bootstrap(ctx context.Context) error {
	// запускаем транзакцию
	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// в случае неуспешного коммита все изменения транзакции будут отменены
	defer tx.Rollback()

	// проверка на существование таблицы
	row := tx.QueryRowContext(ctx, `SELECT EXISTS (
		SELECT 1 FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_name = 'urlstorage'
	) AS table_exists`)

	var tableExists bool

	err = row.Scan(&tableExists)
	if err != nil {
		return err
	}

	if tableExists {
		return tx.Commit()
	}

	// создаём таблицу сообщений и необходимые индексы
	tx.ExecContext(ctx, `
        CREATE TABLE urlstorage (
            shorturl varchar(8) PRIMARY KEY,
            originalurl varchar(512)
        )
    `)
	//tx.ExecContext(ctx, `CREATE INDEX originalurl_idx ON urlstorage (originalurl)`)

	// коммитим транзакцию
	return tx.Commit()
}

func (s *Store) SaveShortURL(ctx context.Context, shortURL, originalURL string) error {
	// ctx := context.Background()
	_, err := s.conn.ExecContext(ctx, `INSERT INTO urlstorage (shortURL, originalURL) VALUES ($1, $2)`, shortURL, originalURL)
	return err
}

func (s *Store) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	row := s.conn.QueryRowContext(ctx, `
		SELECT
			originalurl
		FROM urlstorage
		WHERE shorturl = $1
		`, shortURL)

	var originalURL string
	err := row.Scan(&originalURL)
	if err != nil {
		return "", err
	}
	return originalURL, nil
}

func (s *Store) IsUnique(ctx context.Context, shortURL string) (bool, error) {
	row := s.conn.QueryRowContext(ctx, `
		SELECT
			shorturl
		FROM urlstorage
		WHERE shorturl = $1
		`, shortURL)

	var shortURLDB string
	err := row.Scan(&shortURLDB)
	if err == sql.ErrNoRows {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}

func (s *Store) SaveShortURLs(ctx context.Context, shortOriginalURLs map[string]string) error {
	// ctx := context.Background()
	shortOriginalURLBatch := make(map[string]string)
	i := 0
	for shortURL, originalURL := range shortOriginalURLs {
		if i == 1000 {
			err := s.saveShortURLsBatch(ctx, shortOriginalURLBatch)
			if err != nil {
				return err
			}
			shortOriginalURLBatch = make(map[string]string)
		}
		shortOriginalURLBatch[shortURL] = originalURL
		i++
	}
	err := s.saveShortURLsBatch(ctx, shortOriginalURLBatch)
	return err
}

func (s *Store) saveShortURLsBatch(ctx context.Context, shortOriginalURLBatch map[string]string) error {
	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO urlstorage (shortURL, originalURL) VALUES ($1, $2)")
	if err != nil {
		return err
	}
	for shortURL, originalURL := range shortOriginalURLBatch {
		_, err := stmt.ExecContext(ctx, shortURL, originalURL)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) Ping() error {
	return s.conn.Ping()
}
