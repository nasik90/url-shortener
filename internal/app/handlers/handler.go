// Модуль handler содержит реализацию http методов.
package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
	middleware "github.com/nasik90/url-shortener/internal/app/middlewares"
	"github.com/nasik90/url-shortener/internal/app/storage"
)

// Service - интерфейс, который описывает методы объектов с типом Service
type Service interface {
	GetShortURL(ctx context.Context, originalURL, userID string) (string, error)
	GetOriginalURL(ctx context.Context, shortURL string) (string, error)
	GetShortURLs(ctx context.Context, originalURLs map[string]string, userID string) (map[string]string, error)
	GetUserURLs(ctx context.Context, userID string) (map[string]string, error)
	MarkRecordsForDeletion(ctx context.Context, shortURLs []string, userID string)
	Ping(ctx context.Context) error
	GetURLsStats(ctx context.Context) (int, int, error)
}

// Handler - структура, хранящая объект типа Service.
type Handler struct {
	service       Service
	trustedSubnet string
}

// NewHandler создает экземпляр объекта Handler.
func NewHandler(service Service, trustedSubnet string) *Handler {
	return &Handler{service: service, trustedSubnet: trustedSubnet}
}

// GetShortURL - метод для получения короткого URL по переданному оригинальному URL.
// Оригинальный URL передается в теле запроса.
func (h *Handler) GetShortURL() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var buf bytes.Buffer
		ctx := req.Context()
		userID := middleware.UserIDFromContext(ctx)
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
		status := http.StatusCreated
		shortURL, err := h.service.GetShortURL(ctx, originalURL, userID)
		if err != nil {
			if errors.Is(err, settings.ErrOriginalURLNotUnique) {
				status = http.StatusConflict
			} else {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		res.Header().Set("content-type", "text/plain")
		res.WriteHeader(status)
		res.Write([]byte(shortURL))
	}
}

// GetOriginalURL - метод для получения оригинального URL по переданному короткому URL.
func (h *Handler) GetOriginalURL() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		id := strings.Trim(req.URL.Path, "/")
		originalURL, err := h.service.GetOriginalURL(ctx, id)
		if err != nil {
			if err == storage.ErrRecordMarkedForDel {
				http.Error(res, err.Error(), http.StatusGone)
				return
			}
			http.Error(res, err.Error(), http.StatusNotFound)
			return
		}
		res.Header().Set("Location", originalURL)
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}

// GetShortURLJSON - метод для получения короткого URL по переданному оригинальному URL.
// Оригинальный URL передается в теле запроса в JSON.
func (h *Handler) GetShortURLJSON() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		userID := middleware.UserIDFromContext(ctx)
		var input struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		if input.URL == "" {
			http.Error(res, "empty url", http.StatusBadRequest)
			return
		}

		status := http.StatusCreated
		shortURL, err := h.service.GetShortURL(ctx, input.URL, userID)
		if err != nil {
			if errors.Is(err, settings.ErrOriginalURLNotUnique) {
				status = http.StatusConflict
			} else {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		var output struct {
			Result string `json:"result"`
		}
		output.Result = shortURL
		result, err := json.Marshal(output)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		res.Header().Set("content-type", "application/json")
		res.WriteHeader(status)
		res.Write(result)
	}
}

// Ping - проверяет работоспособность сервера и БД.
func (h *Handler) Ping() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := h.service.Ping(ctx)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// GetShortURLs - принимает на вход массив структур с указанием correlation_id и оригинального URL.
// Возвращает массив струкур с указанием correlation_id и короткого URL.
func (h *Handler) GetShortURLs() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		userID := middleware.UserIDFromContext(ctx)
		type input struct {
			СorrelationID string `json:"correlation_id"`
			OriginalURL   string `json:"original_url"`
		}
		var s []input
		if err := json.NewDecoder(req.Body).Decode(&s); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		originalURLs := make(map[string]string)
		for _, in := range s {
			originalURLs[in.СorrelationID] = in.OriginalURL
		}
		shortURLs, err := h.service.GetShortURLs(ctx, originalURLs, userID)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		type output struct {
			СorrelationID string `json:"correlation_id"`
			ShortURL      string `json:"short_url"`
		}
		var o []output
		for corID, shortURL := range shortURLs {
			o = append(o, output{СorrelationID: corID, ShortURL: shortURL})
		}

		result, err := json.Marshal(o)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		res.Header().Set("content-type", "application/json")
		res.WriteHeader(http.StatusCreated)
		res.Write(result)
	}
}

// GetUserURLs - возвращает список URL`ов пользователя.
// Список представляет собой массив структур с указанием короткого и оригинального URL.
func (h *Handler) GetUserURLs() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		userID := middleware.UserIDFromContext(ctx)
		UserURLs, err := h.service.GetUserURLs(ctx, userID)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		type output struct {
			ShortURL    string `json:"short_url"`
			OriginalURL string `json:"original_url"`
		}
		var o []output
		for shortURL, originalURL := range UserURLs {
			o = append(o, output{ShortURL: shortURL, OriginalURL: originalURL})
		}

		result, err := json.Marshal(o)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		status := http.StatusOK
		if len(UserURLs) == 0 {
			status = http.StatusNoContent
		}

		res.Header().Set("content-type", "application/json")
		res.WriteHeader(status)
		res.Write(result)
	}
}

// MarkRecordsForDeletion помечает на удаление переданные в массиве короткие URL
func (h *Handler) MarkRecordsForDeletion() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var s []string
		ctx := req.Context()
		userID := middleware.UserIDFromContext(ctx)
		if err := json.NewDecoder(req.Body).Decode(&s); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		h.service.MarkRecordsForDeletion(ctx, s, userID)
		res.Header().Set("content-type", "text/plain")
		res.WriteHeader(http.StatusAccepted)
	}
}

// GetUserURLs - возвращает количество URL и пользователей.
// В настройках сервиса обязательно должен быть указан CIDR и передан в заголовке X-Real-IP IP адрес.
func (h *Handler) GetURLsStats() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		urls, users, err := h.service.GetURLsStats(ctx)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		type outputType struct {
			Urls  int `json:"urls"`
			Users int `json:"users"`
		}
		output := outputType{Urls: urls, Users: users}
		result, err := json.Marshal(output)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		status := http.StatusOK
		res.Header().Set("content-type", "application/json")
		res.WriteHeader(status)
		res.Write(result)
	}
}
