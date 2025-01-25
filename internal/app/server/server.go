package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
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
		r.Post("/", getShortURL(repository, &mutex, options.BaseURL))
		r.Post("/api/shorten", getShortURLJSON(repository, &mutex, options.BaseURL))
		r.Get("/{id}", getOriginalURL(repository))
	})
	err := http.ListenAndServe(options.ServerAddress, logger.RequestLogger(r.ServeHTTP))
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
	}
}
