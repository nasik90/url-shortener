package service

import (
	"crypto/rand"
	"math/big"
	"sync"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
	"github.com/nasik90/url-shortener/internal/app/storage"
)

func GetShortURL(repository storage.Repositories, mutex *sync.Mutex, originalURL, host string) (string, error) {
	mutex.Lock()
	shortURL, err := shortURLWithRetrying(repository)
	if err != nil {
		return "", err
	}
	repository.SaveShortURL(shortURL, originalURL)
	mutex.Unlock()

	shortURLWithHost := shortURLWithHost(host, shortURL)
	return shortURLWithHost, nil
}

func GetOriginalURL(repository storage.Repositories, shortURL string) (string, error) {

	originalURL, err := repository.GetOriginalURL(shortURL)
	if err != nil {
		return "", err
	}
	return originalURL, nil

}

func randomString(charCount int) (res string, err error) {

	template := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
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
	// return "http://" + host + "/" + randomString
	return host + "/" + randomString
}

func shortURLWithRetrying(repository storage.Repositories) (string, error) {
	shortURL := ""
	shortURLUnique := false
	for !shortURLUnique {
		randomString, err := randomString(settings.ShortURLlen)
		if err != nil {
			return "", err
		}
		//shortURL = buildShortURL(host, randomString)
		shortURL = randomString
		shortURLUnique = repository.ShortURLUnique(shortURL)
	}
	return shortURL, nil
}
