package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
	"github.com/nasik90/url-shortener/internal/app/logger"
	"github.com/nasik90/url-shortener/internal/app/service"
	"github.com/nasik90/url-shortener/internal/app/storage"
	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
)

type originalURLStruct struct {
	URL string `json:"url"`
}

type shortURLResultStruct struct {
	Result string `json:"result"`
}

func RunServer(repository storage.Repositories, options *settings.Options) {

	var mutex sync.Mutex

	if err := logger.Initialize(options.LogLevel); err != nil {
		panic(err)
	}

	logger.Log.Info("Running server", zap.String("address", options.ServerAddress))

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", (getShortURL(repository, &mutex, options.BaseURL)))
		r.Post("/api/shorten", getShortURLJSON(repository, &mutex, options.BaseURL))
		r.Get("/{id}", getOriginalURL(repository))
	})
	err := http.ListenAndServe(options.ServerAddress, logger.RequestLogger(gzipMiddleware(r.ServeHTTP)))
	if err != nil {
		panic(err)
	}
}

func getShortURL(repository storage.Repositories, mutex *sync.Mutex, host string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var buf bytes.Buffer
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
		shortURL, err := service.GetShortURL(repository, mutex, originalURL, host)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		res.Header().Set("content-type", "text/plain")
		res.Header().Set("Content-Length", strconv.Itoa(len(shortURL)))
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(shortURL))
	}
}

func getOriginalURL(repository storage.Repositories) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		id := req.RequestURI
		//if id with "/"
		if len(id) == settings.ShortURLlen+1 {
			id = id[1:]
		}
		originalURL, err := service.GetOriginalURL(repository, id)
		if err != nil {
			http.Error(res, err.Error(), http.StatusNotFound)
			return
		}
		res.Header().Set("Location", originalURL)
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func getShortURLJSON(repository storage.Repositories, mutex *sync.Mutex, host string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var buf bytes.Buffer
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		var originalURLStruct originalURLStruct
		err = json.Unmarshal(buf.Bytes(), &originalURLStruct)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		if originalURLStruct.URL == "" {
			http.Error(res, "empty url", http.StatusBadRequest)
			return
		}

		shortURL, err := service.GetShortURL(repository, mutex, originalURLStruct.URL, host)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		resultStruct := shortURLResultStruct{Result: shortURL}
		result, err := json.MarshalIndent(resultStruct, "", "    ")
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		res.Header().Set("content-type", "application/json")
		res.Header().Set("Content-Length", strconv.Itoa(len(string(result))))
		res.WriteHeader(http.StatusCreated)
		res.Write(result)
		//logger.Log.Info("Running server", []byte(res))
		// Log.Sugar().Infoln(
		// 	"uri", r.URL.Path,
		// 	"method", r.Method,
		// 	"status", responseData.status,
		// 	"duration", duration,
		// 	"size", responseData.size,
		// )
	}
}

func gzipMiddleware(h http.HandlerFunc) http.HandlerFunc {
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
