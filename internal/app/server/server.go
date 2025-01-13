package server

import (
	"bytes"
	"net/http"
	"strconv"
	"sync"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
	"github.com/nasik90/url-shortener/internal/app/service"
	"github.com/nasik90/url-shortener/internal/app/storage"

	"github.com/go-chi/chi/v5"
)

func RunServer(repository storage.Repositories, mutex *sync.Mutex, options *settings.Options) {

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", getShortURL(repository, mutex, options.B))
		r.Get("/{id}", getOriginalURL(repository))
	})
	err := http.ListenAndServe(options.A, r)
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
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		res.Header().Set("Location", originalURL)
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}
