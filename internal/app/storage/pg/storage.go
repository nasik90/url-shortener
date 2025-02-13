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

// func NewStore(databaseDSN string) (*Store, error) {
// 	conn, err := sql.Open("pgx", databaseDSN)
// 	if err != nil {
// 		return &Store{}, err
// 	}
// 	if conn.Ping() != nil{}
// 	return &Store{conn: conn}, nil
// }

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
	tx.ExecContext(ctx, `CREATE INDEX originalurl_idx ON urlstorage (originalurl)`)

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

func (s *Store) Ping() error {
	return s.conn.Ping()
}
