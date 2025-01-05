package storage

import "errors"

type Repositories interface {
	SaveShortURL(shortURL, originalURL string) error
	GetOriginalURL(shortURL string) (string, error)
	CheckShortURL(shortURL string) error
}

type LocalCache struct {
	CahceMap map[string]string
}

var UrlCache = make(map[string]string)

func (localCache *LocalCache) SaveShortURL(shortURL, originalURL string) error {
	localCache.CahceMap[shortURL] = originalURL
	return nil
}

func (localCache *LocalCache) GetOriginalURL(shortURL string) (string, error) {
	originalURL, ok := localCache.CahceMap[shortURL]
	if !ok {
		err := errors.New("original URL not found")
		return "", err
	}
	return originalURL, nil
}

func (localCache *LocalCache) ShortURLUnique(shortURL string) bool {
	_, ok := localCache.CahceMap[shortURL]
	return !ok
}
