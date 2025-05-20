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
	handler *handler.Handler
}

// NewServer создает экземпляр структуры Server.
func NewServer(handler *handler.Handler, serverAddress string) *Server {
	s := &Server{}
	s.Addr = serverAddress
	s.handler = handler
	return s
}

// RunServer запускает сервер.
func (s *Server) RunServer() error {

	logger.Log.Info("Running server", zap.String("address", s.Addr))

	r := chi.NewRouter()
	//r.Mount("/debug", middleware.Profiler())
	r.Route("/", func(r chi.Router) {
		r.Post("/", s.handler.GetShortURL())
		r.Post("/api/shorten", s.handler.GetShortURLJSON())
		r.Post("/api/shorten/batch", s.handler.GetShortURLs())
		r.Get("/{id}", s.handler.GetOriginalURL())
		r.Get("/ping", s.handler.Ping())
		r.Get("/api/user/urls", s.handler.GetUserURLs())
		r.Delete("/api/user/urls", s.handler.MarkRecordsForDeletion())
	})
	//r.Mount("/debug/pprof", http.HandlerFunc(pprof.Index)) // Главная страница pprof
	//r.Mount("/debug/pprof/", http.DefaultServeMux) // Все маршруты pprof
	s.Handler = logger.RequestLogger(middleware.Auth(middleware.GzipMiddleware(r.ServeHTTP)))
	err := s.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

// RunServer останавливает сервер.
func (s *Server) StopServer() error {
	return s.Shutdown(context.Background())
}
