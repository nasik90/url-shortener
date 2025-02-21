package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
	"github.com/nasik90/url-shortener/internal/app/service"
)

func GetShortURL(repository service.Repository, host string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var buf bytes.Buffer
		ctx := req.Context()
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
		shortURL, err := service.GetShortURL(ctx, repository, originalURL, host)
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

func GetOriginalURL(repository service.Repository) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		id := strings.Trim(req.URL.Path, "/")
		originalURL, err := service.GetOriginalURL(ctx, repository, id)
		if err != nil {
			http.Error(res, err.Error(), http.StatusNotFound)
			return
		}
		res.Header().Set("Location", originalURL)
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func GetShortURLJSON(repository service.Repository, host string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
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
		shortURL, err := service.GetShortURL(ctx, repository, input.URL, host)
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

func Ping(repository service.Repository) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		err := repository.Ping(req.Context())
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func GetShortURLs(repository service.Repository, host string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
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
		shortURLs, err := service.GetShortURLs(ctx, repository, originalURLs, host)
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

		result, err := json.MarshalIndent(o, "", "    ")
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		res.Header().Set("content-type", "application/json")
		res.WriteHeader(http.StatusCreated)
		res.Write(result)
	}
}
