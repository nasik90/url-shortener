package server

import (
	"context"
	"net/http"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
	handler "github.com/nasik90/url-shortener/internal/app/handlers"
	"github.com/nasik90/url-shortener/internal/app/logger"
	middleware "github.com/nasik90/url-shortener/internal/app/middlewares"
	"github.com/nasik90/url-shortener/internal/app/service"
	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	http.Server
	handler *handler.Handler
}

func NewServer(handler *handler.Handler, serverAddress string) *Server {
	s := &Server{}
	s.Addr = serverAddress
	s.handler = handler
	return s
}

func (s *Server) RunServer(repository service.Repository, options *settings.Options) error {

	if err := logger.Initialize(options.LogLevel); err != nil {
		return err
	}

	logger.Log.Info("Running server", zap.String("address", s.Addr))

	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", s.handler.GetShortURL())
		r.Post("/api/shorten", s.handler.GetShortURLJSON())
		r.Post("/api/shorten/batch", s.handler.GetShortURLs())
		r.Get("/{id}", s.handler.GetOriginalURL())
		r.Get("/ping", s.handler.Ping())
		r.Get("/api/user/urls", s.handler.GetUserURLs())
		r.Delete("/api/user/urls", s.handler.MarkRecordsForDeletion())
	})
	s.Handler = logger.RequestLogger(middleware.Auth(middleware.GzipMiddleware(r.ServeHTTP)))
	err := s.ListenAndServe()
	//err := http.ListenAndServe(options.ServerAddress, logger.RequestLogger(middleware.Auth(middleware.GzipMiddleware(r.ServeHTTP))))
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) StopServer() error {
	return s.Shutdown(context.Background())
}
