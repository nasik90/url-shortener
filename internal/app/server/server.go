package server

import (
	"net/http"
	"sync"

	"github.com/nasik90/url-shortener/internal/app/handlers"
	"github.com/nasik90/url-shortener/internal/app/storage"

	"github.com/go-chi/chi/v5"
)

func RunServer(localCache *storage.LocalCache, mutex *sync.Mutex) {

	// mux := http.NewServeMux()
	// mux.HandleFunc("/", handlers.GetShortURL(localCache, mutex))
	// mux.HandleFunc("/{id}", handlers.GetOriginalURL(localCache))
	// err := http.ListenAndServe(`:8080`, mux)
	// if err != nil {
	// 	panic(err)
	// }

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", handlers.GetShortURL(localCache, mutex))
		r.Get("/{id}", handlers.GetOriginalURL(localCache))
	})
	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		panic(err)
	}
}
