package storage

import (
	"context"
	"errors"
	"io"
	"strconv"
	"sync"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
)

var (
	ErrRecordMarkedForDel = errors.New("record marked for deletion")
)

type LocalCache struct {
	mu               sync.RWMutex
	ShortOriginalURL map[string]string
	OriginalShortURL map[string]string
	ShortURLUserID   map[string]string
	MarkedForDelURL  map[string]bool
}

func NewLocalCahce() *LocalCache {
	localCache := &LocalCache{}
	localCache.ShortOriginalURL = make(map[string]string)
	localCache.OriginalShortURL = make(map[string]string)
	localCache.ShortURLUserID = make(map[string]string)
	localCache.MarkedForDelURL = make(map[string]bool)
	return localCache
}

func (l *LocalCache) SaveShortURL(ctx context.Context, shortURL, originalURL, userID string) error {
	l.mu.RLock()
	if _, ok := l.ShortOriginalURL[shortURL]; ok {
		return settings.ErrShortURLNotUnique
	}
	l.mu.RUnlock()
	l.saveShortURL(shortURL, originalURL, userID)
	return nil
}

func (l *LocalCache) saveShortURL(shortURL, originalURL, userID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.ShortOriginalURL[shortURL] = originalURL
	l.OriginalShortURL[originalURL] = shortURL
	l.ShortURLUserID[shortURL] = userID
}

func (l *LocalCache) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	originalURL, ok := l.ShortOriginalURL[shortURL]
	if !ok {
		err := settings.ErrOriginalURLNotFound
		return "", err
	}
	if l.MarkedForDelURL[shortURL] {
		return "", ErrRecordMarkedForDel
	}
	return originalURL, nil
}

func (l *LocalCache) Ping(ctx context.Context) error {
	return nil
}

func (l *LocalCache) Close() error {
	return nil
}

func (l *LocalCache) SaveShortURLs(ctx context.Context, shortOriginalURLs map[string]string, userID string) error {
	for shortURL, originalURL := range shortOriginalURLs {
		err := l.SaveShortURL(ctx, shortURL, originalURL, userID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *LocalCache) GetShortURL(ctx context.Context, originalURL string) (string, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.OriginalShortURL[originalURL], nil
}

func (l *LocalCache) GetUserURLs(ctx context.Context, userID string) (result map[string]string, err error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	result = make(map[string]string)
	for shortURL, savedUserID := range l.ShortURLUserID {
		if savedUserID == userID {
			result[shortURL] = l.ShortOriginalURL[shortURL]
		}
	}
	return result, nil
}

func (l *LocalCache) MarkRecordsForDeletion(ctx context.Context, records ...settings.Record) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, record := range records {
		if l.ShortURLUserID[record.ShortURL] == record.UserID {
			l.MarkedForDelURL[record.ShortURL] = true
		}
	}
	return nil
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
	producer, err := NewProducer(fileName)
	if err != nil {
		return fileStorage, err
	}
	consumer, err := NewConsumer(fileName)
	if err != nil {
		return fileStorage, err
	}
	fileStorage.localCache = NewLocalCahce()
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

func (f *FileStorage) SaveShortURL(ctx context.Context, shortURL, originalURL, userID string) error {
	var event Event
	f.mu.Lock()
	defer f.mu.Unlock()
	f.CurrentUUID++
	event.UUID = strconv.Itoa(f.CurrentUUID)
	event.ShortURL = shortURL
	event.OriginalURL = originalURL
	event.UserID = userID
	if err := f.Producer.WriteEvent(&event); err != nil {
		return err
	}
	f.localCache.saveShortURL(shortURL, originalURL, userID)
	return nil
}

func (f *FileStorage) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	return f.localCache.GetOriginalURL(ctx, shortURL)
}

func restoreData(f *FileStorage) error {
	for {
		event, err := f.Consumer.ReadEvent()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		f.localCache.saveShortURL(event.ShortURL, event.OriginalURL, event.UserID)
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

func (f *FileStorage) SaveShortURLs(ctx context.Context, shortOriginalURLs map[string]string, userID string) error {
	for shortURL, originalURL := range shortOriginalURLs {
		err := f.SaveShortURL(ctx, shortURL, originalURL, userID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *FileStorage) GetShortURL(ctx context.Context, originalURL string) (string, error) {
	return f.localCache.GetShortURL(ctx, originalURL)
}

func (f *FileStorage) GetUserURLs(ctx context.Context, userID string) (result map[string]string, err error) {
	return f.localCache.GetUserURLs(ctx, userID)
}

func (f *FileStorage) MarkRecordsForDeletion(ctx context.Context, records ...settings.Record) error {
	// + Написать обновление записи в файле
	return f.localCache.MarkRecordsForDeletion(ctx, records...)
}
