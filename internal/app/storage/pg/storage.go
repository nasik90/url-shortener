package pg

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
	"github.com/nasik90/url-shortener/internal/app/logger"
	"github.com/nasik90/url-shortener/internal/app/storage"
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
            short_url varchar(8) CONSTRAINT shorturl_pkey PRIMARY KEY NOT NULL,
            original_url varchar(512) CONSTRAINT originalurl_ukey UNIQUE NOT NULL ,
			user_id varchar(64) NOT NULL, 
			deleted_flag bool DEFAULT false NOT NULL  
        )
    `)

	// коммитим транзакцию
	return tx.Commit()
}

func (s *Store) SaveShortURL(ctx context.Context, shortURL, originalURL, userID string) error {
	_, err := s.conn.ExecContext(ctx, `INSERT INTO urlstorage (short_url, original_url, user_id) VALUES ($1, $2, $3)`, shortURL, originalURL, userID)
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
		short_url
	FROM urlstorage
	WHERE original_url = $1
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
			original_url,
			deleted_flag
		FROM urlstorage
		WHERE short_url = $1
		`, shortURL)

	var (
		originalURL string
		deletedFlag bool
	)
	err := row.Scan(&originalURL, &deletedFlag)
	if err != nil {
		return "", err
	}
	if deletedFlag {
		return originalURL, storage.ErrRecordMarkedForDel
	}
	return originalURL, nil
}

func (s *Store) SaveShortURLs(ctx context.Context, shortOriginalURLs map[string]string, userID string) error {
	// при массовом сохранении сейчас нет проверки на уникальность вставляемых shortURL и originalURL
	// как вижу реализацию данной проверки:
	// блокируем строки таблицы по вставляемым shortURL и отдельно по вставляемым originalURL
	// проверяем селектом, что в БД нет таких shortURL и originalURL. Если нет, то вставляем записи.
	// Если есть такие же originalURL, то возвращаем по таким shortURL из БД
	// Если есть такие же shortURL, то по таким записям генерируем новый shortURL
	// Все это делаем в одной транзакции
	const batchLimit = 1000
	shortOriginalURLBatch := make(map[string]string)
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

func (s *Store) saveShortURLsBatch(ctx context.Context, shortOriginalURLBatch map[string]string, userID string) error {
	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO urlstorage (short_url, original_url, user_id) VALUES ($1, $2, $3)")
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

func (s *Store) GetUserURLs(ctx context.Context, userID string) (map[string]string, error) {
	data := make(map[string]string)
	rows, err := s.conn.QueryContext(ctx, `
	SELECT
		short_url,
		original_url
	FROM urlstorage
	WHERE user_id = $1
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

func (s *Store) MarkRecordsForDeletion(ctx context.Context, records ...settings.Record) error {
	for _, r := range records {
		logger.Log.Info("record marked for deletion(plan)", zap.String("shortURL", r.ShortURL), zap.String("userID", r.UserID))
	}

	var shortURLs, userIDs []string
	for _, r := range records {
		shortURLs = append(shortURLs, r.ShortURL)
		userIDs = append(userIDs, r.UserID)
	}

	query := `
		UPDATE urlstorage SET deleted_flag = true FROM unnest($1::text[],$2::text[]) AS input(short_url, user_id) WHERE urlstorage.short_url = input.short_url and urlstorage.user_id = input.user_id 
	`

	res, err := s.conn.ExecContext(ctx, query, shortURLs, userIDs)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	logger.Log.Info("record marked for deletion(fact)", zap.String("rowsAffected", strconv.Itoa(int(rowsAffected))))
	return err
}
