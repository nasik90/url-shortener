package service

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"sync"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
)

type Repositories interface {
	SaveShortURL(ctx context.Context, shortURL, originalURL string) error
	SaveShortURLs(ctx context.Context, shortOriginalURLs map[string]string) error
	GetOriginalURL(ctx context.Context, shortURL string) (string, error)
	IsUnique(ctx context.Context, shortURL string) (bool, error)
	Ping() error
	Close() error
	GetShortURL(ctx context.Context, originalURL string) (string, error)
}

func GetShortURL(ctx context.Context, repository Repositories, mutex *sync.Mutex, originalURL, host string) (string, error) {
	mutex.Lock()
	defer mutex.Unlock()
	shortURL, err := shortURLWithRetrying(ctx, repository)
	if err != nil {
		return "", err
	}
	err = repository.SaveShortURL(ctx, shortURL, originalURL)
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

func GetOriginalURL(ctx context.Context, repository Repositories, shortURL string) (string, error) {
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

func shortURLWithRetrying(ctx context.Context, repository Repositories) (string, error) {
	shortURL := ""
	shortURLUnique := false
	for !shortURLUnique {
		randomString, err := randomString(settings.ShortURLlen)
		if err != nil {
			return "", err
		}
		shortURL = randomString
		shortURLUnique, err = repository.IsUnique(ctx, shortURL)
		if err != nil {
			return "", err
		}
	}
	return shortURL, nil
}

func GetShortURLs(ctx context.Context, repository Repositories, mutex *sync.Mutex, originalURLs map[string]string, host string) (map[string]string, error) {
	mutex.Lock()
	defer mutex.Unlock()

	shortURLs := make(map[string]string)
	shortOriginalURLs := make(map[string]string)

	for id, shortOriginalURL := range originalURLs {
		shortURL, err := shortURLWithRetrying(ctx, repository) // запросы в цикле =(
		if err != nil {
			return shortURLs, err
		}
		shortURLWithHost := shortURLWithHost(host, shortURL)
		shortURLs[id] = shortURLWithHost
		shortOriginalURLs[shortURL] = shortOriginalURL // проверить, что и сгенерированный в данной сессии shortURL уникален
	}
	err := repository.SaveShortURLs(ctx, shortOriginalURLs)
	return shortURLs, err
}
