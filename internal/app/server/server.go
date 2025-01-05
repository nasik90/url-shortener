package server

import (
	"net/http"
	"sync"

	"github.com/nasik90/url-shortener/internal/app/handlers"
	"github.com/nasik90/url-shortener/internal/app/storage"
)

func RunServer(localCache *storage.LocalCache, mutex *sync.Mutex) {

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.GetShortURL(localCache, mutex))
	mux.HandleFunc("/{id}", handlers.GetOriginalURL(localCache))
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}

}
