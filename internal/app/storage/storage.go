package storage

import (
	"github.com/nasik90/url-shortener/cmd/shortener/settings"
)

// type Repositories interface {
// 	SaveShortURL(shortURL, originalURL string) error
// 	GetOriginalURL(shortURL string) (string, error)
// 	ShortURLUnique(shortURL string) bool
// }

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
		err := settings.ErrOriginalURLNotFound
		return "", err
	}
	return originalURL, nil
}

func (localCache *LocalCache) ShortURLUnique(shortURL string) bool {
	_, ok := localCache.CahceMap[shortURL]
	return !ok
}

// func (localCache *LocalCache) RestoreData(filePath string) error {
// 	consumer, err := NewConsumer(filePath)
// 	if err != nil {
// 		return err
// 	}
// 	for {
// 		event, err := consumer.ReadEvent()
// 		if err != nil {
// 			if err == io.EOF {
// 				break
// 			}
// 			return err
// 		}
// 		localCache.SaveShortURL(event.ShortURL, event.OriginalURL)
// 		CurrentUUID, err = strconv.Atoi(event.UUID)
// 		if err != nil {
// 			return nil
// 		}
// 	}

// 	consumer.Close()
// 	return nil
// }

// func (localCache *LocalCache) WriteDataToFile()(

// )
