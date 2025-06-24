// Модуль server служит для запуска сервера с указанием http методов.
package server

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	handler "github.com/nasik90/url-shortener/internal/app/handlers"
	"github.com/nasik90/url-shortener/internal/app/logger"
	middleware "github.com/nasik90/url-shortener/internal/app/middlewares"

	"github.com/go-chi/chi/v5"
)

// Server - структура, которая характеризует сервер.
// Содержит встроенную структуру из типовой библиотеки http.Server и handler.
type Server struct {
	http.Server
	handler     *handler.Handler
	enableHTTPS bool
}

// NewServer создает экземпляр структуры Server.
func NewServer(handler *handler.Handler, serverAddress string, enableHTTPS bool) *Server {
	s := &Server{}
	s.Addr = serverAddress
	s.handler = handler
	s.enableHTTPS = enableHTTPS
	return s
}

// RunServer запускает сервер.
func (s *Server) RunServer() error {

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
	var err error
	if s.enableHTTPS {
		err = s.ListenAndServeTLS("server.crt", "server.key")
	} else {
		err = s.ListenAndServe()
	}
	if err != nil {
		return err
	}

	return nil
}

// RunServer останавливает сервер.
func (s *Server) StopServer(ctx context.Context) error {
	return s.Shutdown(ctx)
}
