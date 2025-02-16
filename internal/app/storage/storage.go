package storage

import (
	"context"
	"io"
	"strconv"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
)

type LocalCache struct {
	CahceMap map[string]string
}

func (localCache *LocalCache) SaveShortURL(ctx context.Context, shortURL, originalURL string) error {
	localCache.CahceMap[shortURL] = originalURL
	return nil
}

func (localCache *LocalCache) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	originalURL, ok := localCache.CahceMap[shortURL]
	if !ok {
		err := settings.ErrOriginalURLNotFound
		return "", err
	}
	return originalURL, nil
}

func (localCache *LocalCache) IsUnique(ctx context.Context, shortURL string) (bool, error) {
	_, ok := localCache.CahceMap[shortURL]
	return !ok, nil
}

func (localCache *LocalCache) Ping() error {
	return nil
}

func (localCache *LocalCache) Close() error {
	return nil
}

func (localCache *LocalCache) SaveShortURLs(ctx context.Context, shortOriginalURLs map[string]string) error {
	for shortURL, originalURL := range shortOriginalURLs {
		err := localCache.SaveShortURL(ctx, shortURL, originalURL)
		if err != nil {
			return err
		}
	}
	return nil
}

func (localCache *LocalCache) GetShortURL(ctx context.Context, originalURL string) (string, error) {
	return "", nil
}

type FileStorage struct {
	localCache  *LocalCache
	CurrentUUID int
	Producer    *Producer
	Consumer    *Consumer
}

func NewFileStorage(fileName string) (*FileStorage, error) {
	fileStorage := &FileStorage{}
	cache := make(map[string]string)
	producer, err := NewProducer(fileName)
	if err != nil {
		return fileStorage, err
	}
	consumer, err := NewConsumer(fileName)
	if err != nil {
		return fileStorage, err
	}
	fileStorage.localCache = &LocalCache{CahceMap: cache}
	fileStorage.Consumer = consumer
	fileStorage.Producer = producer
	err = restoreData(fileStorage)
	if err != nil {
		return fileStorage, err
	}
	return fileStorage, nil
}

func (fileStorage *FileStorage) Close() error {
	return fileStorage.Producer.Close()
}

func (fileStorage *FileStorage) SaveShortURL(ctx context.Context, shortURL, originalURL string) error {
	var event Event
	fileStorage.CurrentUUID++
	event.UUID = strconv.Itoa(fileStorage.CurrentUUID)
	event.ShortURL = shortURL
	event.OriginalURL = originalURL
	fileStorage.Producer.WriteEvent(&event)
	return fileStorage.localCache.SaveShortURL(ctx, shortURL, originalURL)
}

func (fileStorage *FileStorage) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	return fileStorage.localCache.GetOriginalURL(ctx, shortURL)
}

func (fileStorage *FileStorage) IsUnique(ctx context.Context, shortURL string) (bool, error) {
	return fileStorage.localCache.IsUnique(ctx, shortURL)
}

func restoreData(fileStorage *FileStorage) error {
	ctx := context.Background()
	for {
		event, err := fileStorage.Consumer.ReadEvent()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		err = fileStorage.localCache.SaveShortURL(ctx, event.ShortURL, event.OriginalURL)
		if err != nil {
			return err
		}
		fileStorage.CurrentUUID, err = strconv.Atoi(event.UUID)
		if err != nil {
			return err
		}
	}

	fileStorage.Consumer.Close()
	return nil
}

func (fileStorage *FileStorage) Ping() error {
	return nil
}

func (fileStorage *FileStorage) SaveShortURLs(ctx context.Context, shortOriginalURLs map[string]string) error {
	for shortURL, originalURL := range shortOriginalURLs {
		err := fileStorage.SaveShortURL(ctx, shortURL, originalURL)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fileStorage *FileStorage) GetShortURL(ctx context.Context, originalURL string) (string, error) {
	return "", nil
}
