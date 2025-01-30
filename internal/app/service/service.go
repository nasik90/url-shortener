package service

import (
	"crypto/rand"
	"math/big"
	"strconv"
	"sync"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
	"github.com/nasik90/url-shortener/internal/app/storage"
)

func GetShortURL(repository storage.Repositories, mutex *sync.Mutex, originalURL, host string) (string, error) {
	var event storage.Event
	mutex.Lock()
	shortURL, err := shortURLWithRetrying(repository)
	if err != nil {
		return "", err
	}
	repository.SaveShortURL(shortURL, originalURL)
	storage.CurrentUUID++
	event.UUID = strconv.Itoa(storage.CurrentUUID)
	event.ShortURL = shortURL
	event.OriginalURL = originalURL
	storage.URLWriterTiFile.WriteEvent(&event)
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

func RestoreData(repository storage.Repositories, filePath string) error {
	return repository.RestoreData(filePath)
}
