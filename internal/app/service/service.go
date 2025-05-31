// Модуль service служит для реализации бизнес логики.
package service

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
)

// Интерфейс Repository описывает методы типа Repository.
type Repository interface {
	SaveShortURL(ctx context.Context, shortURL, originalURL, userID string) error
	SaveShortURLs(ctx context.Context, shortOriginalURLs map[string]string, userID string) error
	GetOriginalURL(ctx context.Context, shortURL string) (string, error)
	Ping(ctx context.Context) error
	Close() error
	GetShortURL(ctx context.Context, originalURL string) (string, error)
	GetUserURLs(ctx context.Context, userID string) (map[string]string, error)
	MarkRecordsForDeletion(ctx context.Context, records ...settings.Record) error
}

// Service - структура, которая хранит ссылку на репозиторий, адрес хоста и канал для хранения URL`ов к удалению.
type Service struct {
	repo          Repository
	host          string
	recordsForDel chan settings.Record
}

// NewService создает экземпляр объекта типа Service.
func NewService(store Repository, host string) *Service {
	return &Service{repo: store, host: host, recordsForDel: make(chan settings.Record)}
}

// GetShortURL - реализует логику по получению короткой ссылки по оригинальной.
func (s *Service) GetShortURL(ctx context.Context, originalURL, userID string) (string, error) {
	shortURL, err := newShortURL()
	if err != nil {
		return "", err
	}
	err = s.repo.SaveShortURL(ctx, shortURL, originalURL, userID)
	for errors.Is(err, settings.ErrShortURLNotUnique) {
		err = s.repo.SaveShortURL(ctx, shortURL, originalURL, userID)
	}

	if err != nil {
		if errors.Is(err, settings.ErrOriginalURLNotUnique) {
			shortURL, err = s.repo.GetShortURL(ctx, originalURL)
			if err != nil {
				return "", err
			}
			shortURLWithHost := shortURLWithHost(s.host, shortURL)
			return shortURLWithHost, settings.ErrOriginalURLNotUnique
		} else {
			return "", err
		}
	}

	shortURLWithHost := shortURLWithHost(s.host, shortURL)
	return shortURLWithHost, nil
}

// GetOriginalURL - реализует логику по получению оригинальной ссылки по короткому
func (s *Service) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	originalURL, err := s.repo.GetOriginalURL(ctx, shortURL)
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

func newShortURL() (string, error) {
	return randomString(settings.ShortURLlen)
}

// GetShortURLs - реализует логику по получению списка коротких ссылок по переданным коротким.
// На входе принимает мапу, где ключ - id, значение - оригинальный урл.
// На выходе тот же id, значение - короткий урл.
func (s *Service) GetShortURLs(ctx context.Context, originalURLs map[string]string, userID string) (map[string]string, error) {
	shortURLs := make(map[string]string)
	shortOriginalURLs := make(map[string]string)

	for id, shortOriginalURL := range originalURLs {
		shortURL, err := newShortURL()
		if err != nil {
			return shortURLs, err
		}
		shortURLWithHost := shortURLWithHost(s.host, shortURL)
		shortURLs[id] = shortURLWithHost
		shortOriginalURLs[shortURL] = shortOriginalURL
	}
	err := s.repo.SaveShortURLs(ctx, shortOriginalURLs, userID)
	return shortURLs, err
}

// GetUserURLs - реализует логику получения списка коротких и оригинальных урлов пользователя.
// На входе id пользователя, на выходе мапа(ключ - короткий урл, значение - оригинальный).
func (s *Service) GetUserURLs(ctx context.Context, userID string) (map[string]string, error) {
	data := make(map[string]string)
	userURLs, err := s.repo.GetUserURLs(ctx, userID)
	if err != nil {
		return data, err
	}
	for shortURL, originalURL := range userURLs {
		shortURLWithHost := shortURLWithHost(s.host, shortURL)
		data[shortURLWithHost] = originalURL
	}
	return data, err
}

// MarkRecordsForDeletion - реализует логику пометки на удаление переданный коротких урл пользователя.
// В данном методе короткие урлы помещаются в канал recordsForDel.
func (s *Service) MarkRecordsForDeletion(ctx context.Context, shortURLs []string, userID string) {
	for _, shortURL := range shortURLs {
		r := settings.Record{
			ShortURL: shortURL,
			UserID:   userID,
		}
		s.recordsForDel <- r
	}
}

// Ping - пингует БД.
func (s *Service) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}
