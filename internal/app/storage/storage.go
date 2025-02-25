package storage

import (
	"context"
	"io"
	"strconv"
	"sync"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
)

type LocalCache struct {
	mu       sync.RWMutex
	CahceMap map[string]string
}

func (l *LocalCache) SaveShortURL(ctx context.Context, shortURL, originalURL string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.CahceMap[shortURL]; ok {
		return settings.ErrShortURLNotUnique
	}
	l.CahceMap[shortURL] = originalURL
	return nil
}

func (l *LocalCache) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	originalURL, ok := l.CahceMap[shortURL]
	if !ok {
		err := settings.ErrOriginalURLNotFound
		return "", err
	}
	return originalURL, nil
}

func (l *LocalCache) Ping(ctx context.Context) error {
	return nil
}

func (l *LocalCache) Close() error {
	return nil
}

func (l *LocalCache) SaveShortURLs(ctx context.Context, shortOriginalURLs map[string]string) error {
	for shortURL, originalURL := range shortOriginalURLs {
		err := l.SaveShortURL(ctx, shortURL, originalURL)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *LocalCache) GetShortURL(ctx context.Context, originalURL string) (string, error) {
	return "", nil
}

func (l *LocalCache) GetUserURLs(ctx context.Context) (result map[string]string, err error) {
	return result, nil
}

type FileStorage struct {
	mu          sync.RWMutex
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

func (f *FileStorage) Close() error {
	return f.Producer.Close()
}

func (f *FileStorage) SaveShortURL(ctx context.Context, shortURL, originalURL string) error {
	var event Event
	f.mu.Lock()
	defer f.mu.Unlock()
	f.CurrentUUID++
	event.UUID = strconv.Itoa(f.CurrentUUID)
	event.ShortURL = shortURL
	event.OriginalURL = originalURL
	if err := f.Producer.WriteEvent(&event); err != nil {
		return err
	}
	return f.localCache.SaveShortURL(ctx, shortURL, originalURL)
}

func (f *FileStorage) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	return f.localCache.GetOriginalURL(ctx, shortURL)
}

func restoreData(f *FileStorage) error {
	ctx := context.Background()
	for {
		event, err := f.Consumer.ReadEvent()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		err = f.localCache.SaveShortURL(ctx, event.ShortURL, event.OriginalURL)
		if err != nil {
			return err
		}
		f.CurrentUUID, err = strconv.Atoi(event.UUID)
		if err != nil {
			return err
		}
	}

	f.Consumer.Close()
	return nil
}

func (f *FileStorage) Ping(ctx context.Context) error {
	return nil
}

func (f *FileStorage) SaveShortURLs(ctx context.Context, shortOriginalURLs map[string]string) error {
	for shortURL, originalURL := range shortOriginalURLs {
		err := f.SaveShortURL(ctx, shortURL, originalURL)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *FileStorage) GetShortURL(ctx context.Context, originalURL string) (string, error) {
	return "", nil
}

func (f *FileStorage) GetUserURLs(ctx context.Context) (result map[string]string, err error) {
	return result, nil
}
