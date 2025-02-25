package server

import (
	"net/http"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
	handler "github.com/nasik90/url-shortener/internal/app/handlers"
	"github.com/nasik90/url-shortener/internal/app/logger"
	middleware "github.com/nasik90/url-shortener/internal/app/middlewares"
	"github.com/nasik90/url-shortener/internal/app/service"
	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
)

func RunServer(repository service.Repository, options *settings.Options) error {

	if err := logger.Initialize(options.LogLevel); err != nil {
		return err
	}

	logger.Log.Info("Running server", zap.String("address", options.ServerAddress))

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", handler.GetShortURL(repository, options.BaseURL))
		r.Post("/api/shorten", handler.GetShortURLJSON(repository, options.BaseURL))
		r.Post("/api/shorten/batch", handler.GetShortURLs(repository, options.BaseURL))
		r.Get("/{id}", handler.GetOriginalURL(repository))
		r.Get("/ping", handler.Ping(repository))
		r.Post("/api/user/urls", handler.GetUserURLs(repository, options.BaseURL))
	})
	err := http.ListenAndServe(options.ServerAddress, logger.RequestLogger(middleware.Auth(middleware.GzipMiddleware(r.ServeHTTP))))
	if err != nil {
		return err
	}

	return nil

}
