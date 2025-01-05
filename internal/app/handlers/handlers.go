package handlers

import (
	"bytes"
	"crypto/rand"
	"math/big"
	"net/http"
	"strconv"
	"sync"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
	"github.com/nasik90/url-shortener/internal/app/storage"
)

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

func buildShortURL(host, randomString string) string {
	return "http://" + host + "/" + randomString
}

func shortURLWithRetrying(host string, localCache *storage.LocalCache) (string, error) {
	shortURL := ""
	shortURLUnique := false
	for !shortURLUnique {
		randomString, err := randomString(settings.ShortURLlen)
		if err != nil {
			return "", err
		}
		shortURL = buildShortURL(host, randomString)
		shortURLUnique = localCache.ShortURLUnique(shortURL)
	}
	return shortURL, nil
}

func GetShortURL(localCache *storage.LocalCache, mutex *sync.Mutex) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var buf bytes.Buffer
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		url := buf.String()
		mutex.Lock()
		shortURL, err := shortURLWithRetrying(req.Host, localCache)
		mutex.Unlock()
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		localCache.SaveShortURL(shortURL, url)
		res.Header().Set("content-type", "text/plain")
		res.Header().Set("Content-Length", strconv.Itoa(len(shortURL)))
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(shortURL))
	}
}

func GetOriginalURL(localCache *storage.LocalCache) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		id := req.RequestURI
		if len(id) == settings.ShortURLlen+1 {
			id = id[1:]
		}
		shortURL := buildShortURL(req.Host, id)
		originalURL, err := localCache.GetOriginalURL(shortURL)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		res.Header().Set("Location", originalURL)
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}
