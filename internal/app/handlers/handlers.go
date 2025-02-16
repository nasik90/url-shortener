package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
	"github.com/nasik90/url-shortener/internal/app/service"
)

// func (m *MyStruct) GetShortURL(res http.ResponseWriter, req *http.Request){

// }

func GetShortURL(repository service.Repositories, mutex *sync.Mutex, host string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var buf bytes.Buffer
		ctx := req.Context()
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		originalURL := buf.String()
		if originalURL == "" {
			http.Error(res, "empty url", http.StatusBadRequest)
			return
		}
		shortURL, err := service.GetShortURL(ctx, repository, mutex, originalURL, host)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.Header().Set("content-type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(shortURL))
	}
}

func GetOriginalURL(repository service.Repositories) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		id := req.RequestURI
		//if id with "/"
		if len(id) == settings.ShortURLlen+1 {
			id = id[1:]
		}
		originalURL, err := service.GetOriginalURL(ctx, repository, id)
		if err != nil {
			http.Error(res, err.Error(), http.StatusNotFound)
			return
		}
		res.Header().Set("Location", originalURL)
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func GetShortURLJSON(repository service.Repositories, mutex *sync.Mutex, host string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		var input struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		if input.URL == "" {
			http.Error(res, "empty url", http.StatusBadRequest)
			return
		}

		shortURL, err := service.GetShortURL(ctx, repository, mutex, input.URL, host)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		var output struct {
			Result string `json:"result"`
		}
		output.Result = shortURL
		result, err := json.MarshalIndent(output, "", "    ")
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		res.Header().Set("content-type", "application/json")
		res.WriteHeader(http.StatusCreated)
		res.Write(result)
	}
}

func GzipMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
		// который будем передавать следующей функции
		ow := w

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			cw := newCompressWriter(w)
			// меняем оригинальный http.ResponseWriter на новый
			ow = cw
			// ow.Header().Set("Content-Encoding", "gzip")
			// не забываем отправить клиенту все сжатые данные после завершения middleware
			defer cw.Close()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			r.Body = cr
			defer cr.Close()
		}

		// передаём управление хендлеру
		h.ServeHTTP(ow, r)
	}
}

func Ping(repository service.Repositories) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		err := repository.Ping()
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func GetShortURLs(repository service.Repositories, mutex *sync.Mutex, host string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		type input struct {
			СorrelationId string `json:"correlation_id"`
			OriginalURL   string `json:"original_url"`
		}
		var s []input
		if err := json.NewDecoder(req.Body).Decode(&s); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		originalURLs := make(map[string]string)
		for _, in := range s {
			originalURLs[in.СorrelationId] = in.OriginalURL
		}
		shortURLs, err := service.GetShortURLs(ctx, repository, mutex, originalURLs, host)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println(shortURLs)
		type output struct {
			СorrelationId string `json:"correlation_id"`
			ShortURL      string `json:"short_url"`
		}
		var o []output
		for corId, shortURL := range shortURLs {
			o = append(o, output{СorrelationId: corId, ShortURL: shortURL})
		}

		result, err := json.MarshalIndent(o, "", "    ")
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		res.Header().Set("content-type", "application/json")
		res.WriteHeader(http.StatusCreated)
		res.Write(result)
	}
}
