package service

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
	"github.com/nasik90/url-shortener/internal/app/logger"
	"go.uber.org/zap"
)

type Repository interface {
	SaveShortURL(ctx context.Context, shortURL, originalURL string) error
	SaveShortURLs(ctx context.Context, shortOriginalURLs map[string]string) error
	GetOriginalURL(ctx context.Context, shortURL string) (string, error)
	Ping(ctx context.Context) error
	Close() error
	GetShortURL(ctx context.Context, originalURL string) (string, error)
	GetUserURLs(ctx context.Context) (map[string]string, error)
	MarkRecordsForDeletion(ctx context.Context, records ...settings.Record) error
}

func GetShortURL(ctx context.Context, repository Repository, originalURL, host string) (string, error) {
	shortURL, err := newShortURL(ctx)
	if err != nil {
		return "", err
	}
	err = repository.SaveShortURL(ctx, shortURL, originalURL)
	for errors.Is(err, settings.ErrShortURLNotUnique) {
		err = repository.SaveShortURL(ctx, shortURL, originalURL)
	}

	if err != nil {
		if errors.Is(err, settings.ErrOriginalURLNotUnique) {
			shortURL, err = repository.GetShortURL(ctx, originalURL)
			if err != nil {
				return "", err
			}
			shortURLWithHost := shortURLWithHost(host, shortURL)
			return shortURLWithHost, settings.ErrOriginalURLNotUnique
		} else {
			return "", err
		}
	}

	shortURLWithHost := shortURLWithHost(host, shortURL)
	return shortURLWithHost, nil
}

func GetOriginalURL(ctx context.Context, repository Repository, shortURL string) (string, error) {
	originalURL, err := repository.GetOriginalURL(ctx, shortURL)
	if err != nil {
		return "", err
	}
	return originalURL, nil
}

func randomString(charCount int) (res string, err error) {
	template := []rune(settings.TemplateForRand)
	templateLen := len(template)
	resChar := make([]rune, charCount)
	for i := 0; i < charCount; i++ {

		r, err := rand.Int(rand.Reader, big.NewInt(int64(templateLen)))
		if err != nil {
			return res, err
		}
		resChar[i] = template[int(r.Int64())]
	}

	return string(resChar), nil
}

func shortURLWithHost(host, randomString string) string {
	return host + "/" + randomString
}

func newShortURL(ctx context.Context) (string, error) {
	return randomString(settings.ShortURLlen)
}

func GetShortURLs(ctx context.Context, repository Repository, originalURLs map[string]string, host string) (map[string]string, error) {
	shortURLs := make(map[string]string)
	shortOriginalURLs := make(map[string]string)

	for id, shortOriginalURL := range originalURLs {
		shortURL, err := newShortURL(ctx)
		if err != nil {
			return shortURLs, err
		}
		shortURLWithHost := shortURLWithHost(host, shortURL)
		shortURLs[id] = shortURLWithHost
		shortOriginalURLs[shortURL] = shortOriginalURL
	}
	err := repository.SaveShortURLs(ctx, shortOriginalURLs)
	return shortURLs, err
}

func GetUserURLs(ctx context.Context, repository Repository, host string) (map[string]string, error) {
	data := make(map[string]string)
	userURLs, err := repository.GetUserURLs(ctx)
	if err != nil {
		return data, err
	}
	for shortURL, originalURL := range userURLs {
		shortURLWithHost := shortURLWithHost(host, shortURL)
		data[shortURLWithHost] = originalURL
	}
	return data, err
}

func MarkRecordsForDeletion(ctx context.Context, repository Repository, shortURLs []string, ch chan<- settings.Record) {
	userID := userIDFromContext(ctx)
	for _, shortURL := range shortURLs {
		var r settings.Record
		r.ShortURL = shortURL
		r.UserID = userID
		ch <- r
	}
}

func userIDFromContext(ctx context.Context) string {
	return ctx.Value(settings.UserIDContextKey).(string)
}

func HandleRecords(repository Repository, ch <-chan settings.Record) {
	// будем сохранять сообщения, накопленные за последние 5 секунд
	ticker := time.NewTicker(5 * time.Second)

	var records []settings.Record
	for {
		select {
		case record := <-ch:
			// добавим сообщение в слайс для последующего сохранения
			records = append(records, record)
		case <-ticker.C:
			// подождём, пока придёт хотя бы одно сообщение
			if len(records) == 0 {
				continue
			}
			// сохраним все пришедшие сообщения одновременно
			err := repository.MarkRecordsForDeletion(context.TODO(), records...)
			if err != nil {
				logger.Log.Info("cannot mark records for deletion", zap.Error(err))
				// не будем стирать сообщения, попробуем отправить их чуть позже
				continue
			}
			// сотрём успешно отосланные сообщения
			records = nil
		}
	}
}
