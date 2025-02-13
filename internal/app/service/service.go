package service

import (
	"context"
	"crypto/rand"
	"math/big"
	"sync"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
)

type Repositories interface {
	SaveShortURL(ctx context.Context, shortURL, originalURL string) error
	GetOriginalURL(ctx context.Context, shortURL string) (string, error)
	IsUnique(ctx context.Context, shortURL string) (bool, error)
	Ping() error
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
		return "", err
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
