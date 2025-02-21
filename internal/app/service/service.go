package service

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
)

type Repository interface {
	SaveShortURL(ctx context.Context, shortURL, originalURL string) error
	SaveShortURLs(ctx context.Context, shortOriginalURLs map[string]string) error
	GetOriginalURL(ctx context.Context, shortURL string) (string, error)
	Ping(ctx context.Context) error
	Close() error
	GetShortURL(ctx context.Context, originalURL string) (string, error)
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
	shortOriginalURLs, shortURLs, err := newShortURLs(ctx, originalURLs, host)
	err = repository.SaveShortURLs(ctx, shortOriginalURLs)
	return shortURLs, err
}

func newShortURLs(ctx context.Context, originalURLs map[string]string, host string) (map[string]string, map[string]string, error) {
	shortURLs := make(map[string]string)
	shortOriginalURLs := make(map[string]string)

	for id, shortOriginalURL := range originalURLs {
		shortURL, err := newShortURL(ctx)
		if err != nil {
			return shortOriginalURLs, shortURLs, err
		}
		shortURLWithHost := shortURLWithHost(host, shortURL)
		shortURLs[id] = shortURLWithHost
		shortOriginalURLs[shortURL] = shortOriginalURL
	}
	return shortOriginalURLs, shortURLs, nil
}
