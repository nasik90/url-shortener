package server

import (
	"net/http"
	"sync"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
	"github.com/nasik90/url-shortener/internal/app/handlers"
	"github.com/nasik90/url-shortener/internal/app/logger"
	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
)

// type shortURLResultStruct struct {
// 	Result string `json:"result"`
// }

func RunServer(repository handlers.Repositories, options *settings.Options) error {

	var mutex sync.Mutex
	if err := logger.Initialize(options.LogLevel); err != nil {
		return err
	}

	logger.Log.Info("Running server", zap.String("address", options.ServerAddress))

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", (handlers.GetShortURL(repository, &mutex, options.BaseURL)))
		r.Post("/api/shorten", handlers.GetShortURLJSON(repository, &mutex, options.BaseURL))
		r.Get("/{id}", handlers.GetOriginalURL(repository))
	})
	err := http.ListenAndServe(options.ServerAddress, logger.RequestLogger(handlers.GzipMiddleware(r.ServeHTTP)))
	if err != nil {
		return err
	}

	return nil

}
