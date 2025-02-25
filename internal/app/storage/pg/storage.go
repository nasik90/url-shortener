package pg

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nasik90/url-shortener/cmd/shortener/settings"
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

	// создаём таблицу сообщений и необходимые индексы
	tx.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS urlstorage (
            shorturl varchar(8) CONSTRAINT shorturl_pkey PRIMARY KEY,
            originalurl varchar(512) CONSTRAINT originalurl_ukey UNIQUE,
			userid bigint
        )
    `)

	tx.ExecContext(ctx, `
        CREATE INDEX IF NOT EXISTS userid_idx ON urlstorage(userID)
    `)

	// коммитим транзакцию
	return tx.Commit()
}

func (s *Store) SaveShortURL(ctx context.Context, shortURL, originalURL string) error {
	userID := userIDFromContext(ctx)
	_, err := s.conn.ExecContext(ctx, `INSERT INTO urlstorage (shortURL, originalURL, userid) VALUES ($1, $2, $3)`, shortURL, originalURL, userID)
	err = checkInsertError(err)
	return err
}

func checkInsertError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		if pgErr.ConstraintName == "shorturl_pkey" {
			return settings.ErrShortURLNotUnique
		}
		if pgErr.ConstraintName == "originalurl_ukey" {
			return settings.ErrOriginalURLNotUnique
		}
	}
	return err
}

func (s *Store) GetShortURL(ctx context.Context, originalURL string) (string, error) {

	row := s.conn.QueryRowContext(ctx, `
	SELECT
		shorturl
	FROM urlstorage
	WHERE originalurl = $1
	`, originalURL)

	var shortURL string
	err := row.Scan(&shortURL)
	if err != nil {
		return "", err
	}
	return shortURL, nil

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

func (s *Store) SaveShortURLs(ctx context.Context, shortOriginalURLs map[string]string) error {
	// при массовом сохранении сейчас нет проверки на уникальность вставляемых shortURL и originalURL
	// как вижу реализацию данной проверки:
	// блокируем строки таблицы по вставляемым shortURL и отдельно по вставляемым originalURL
	// проверяем селектом, что в БД нет таких shortURL и originalURL. Если нет, то вставляем записи.
	// Если есть такие же originalURL, то возвращаем по таким shortURL из БД
	// Если есть такие же shortURL, то по таким записям генерируем новый shortURL
	// Все это делаем в одной транзакции
	const batchLimit = 1000
	shortOriginalURLBatch := make(map[string]string)
	userID := userIDFromContext(ctx)
	i := 0
	for shortURL, originalURL := range shortOriginalURLs {
		if i == batchLimit {
			err := s.saveShortURLsBatch(ctx, shortOriginalURLBatch, userID)
			if err != nil {
				return err
			}
			for k := range shortOriginalURLBatch {
				delete(shortOriginalURLBatch, k)
			}
			i = 0
		}
		shortOriginalURLBatch[shortURL] = originalURL
		i++
	}
	err := s.saveShortURLsBatch(ctx, shortOriginalURLBatch, userID)
	return err
}

func (s *Store) saveShortURLsBatch(ctx context.Context, shortOriginalURLBatch map[string]string, userID int) error {
	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO urlstorage (shortURL, originalURL, userid) VALUES ($1, $2, $3)")
	if err != nil {
		return err
	}
	for shortURL, originalURL := range shortOriginalURLBatch {
		_, err := stmt.ExecContext(ctx, shortURL, originalURL, userID)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) Ping(ctx context.Context) error {
	return s.conn.PingContext(ctx)
}

func (s *Store) GetUserURLs(ctx context.Context) (map[string]string, error) {
	userID := userIDFromContext(ctx)
	data := make(map[string]string)
	rows, err := s.conn.QueryContext(ctx, `
	SELECT
		shorturl,
		originalurl
	FROM urlstorage
	WHERE userid = $1
	`, userID)

	if err != nil {
		return data, err
	}

	var (
		shortURL, originalURL string
	)
	for rows.Next() {
		if err := rows.Scan(&shortURL, &originalURL); err != nil {
			return data, err
		}
		data[shortURL] = originalURL
	}

	if err := rows.Err(); err != nil {
		return data, err
	}

	return data, nil
}

func userIDFromContext(ctx context.Context) int {
	return ctx.Value(settings.ContextUserIDKey).(int)
}
