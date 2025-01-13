package storage

import (
	"errors"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
)

type Repositories interface {
	SaveShortURL(shortURL, originalURL string) error
	GetOriginalURL(shortURL string) (string, error)
	ShortURLUnique(shortURL string) bool
}

type LocalCache struct {
	CahceMap map[string]string
}

func (localCache *LocalCache) SaveShortURL(shortURL, originalURL string) error {
	localCache.CahceMap[shortURL] = originalURL
	return nil
}

func (localCache *LocalCache) GetOriginalURL(shortURL string) (string, error) {
	originalURL, ok := localCache.CahceMap[shortURL]
	if !ok {
		err := errors.New(settings.OriginalURLNotFoundErr)
		return "", err
	}
	return originalURL, nil
}

func (localCache *LocalCache) ShortURLUnique(shortURL string) bool {
	_, ok := localCache.CahceMap[shortURL]
	return !ok
}
