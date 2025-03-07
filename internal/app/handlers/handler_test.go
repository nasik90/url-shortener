package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
	middleware "github.com/nasik90/url-shortener/internal/app/middlewares"
	"github.com/nasik90/url-shortener/internal/app/service"
	"github.com/nasik90/url-shortener/internal/app/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetShortURL(t *testing.T) {
	ctx := context.Background()
	repo := storage.NewLocalCahce()
	type want struct {
		code              int
		originalURLFromDB string
	}
	tests := []struct {
		name        string
		originalURL string
		userID      string
		want        want
	}{
		{
			name:        "positive test #1",
			originalURL: "https://practicum.yandex.ru/",
			userID:      "123",
			want: want{
				code:              http.StatusCreated,
				originalURLFromDB: "https://practicum.yandex.ru/",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := httptest.NewRecorder().Body
			body.Write([]byte(tt.originalURL))
			request := httptest.NewRequest(http.MethodPost, "/", body).
				WithContext(context.WithValue(context.Background(), middleware.UserIDContextKey{}, tt.userID))
			w := httptest.NewRecorder()

			service := service.NewService(repo, request.Host)
			handler := NewHandler(service)

			handler.GetShortURL()(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			resBodyString := string(resBody)
			shortURL := string(resBody)[len(resBodyString)-settings.ShortURLlen:]
			originalURLFromDB, err := repo.GetOriginalURL(ctx, shortURL)
			require.NoError(t, err)
			assert.Equal(t, tt.want.originalURLFromDB, originalURLFromDB)
		})
	}
}

func TestGetOriginalURL(t *testing.T) {
	ctx := context.Background()
	repo := storage.NewLocalCahce()
	type want struct {
		code         int
		responseBody string
		location     string
	}
	tests := []struct {
		name        string
		shortURL    string
		originalURL string
		userID      string
		want        want
	}{
		{
			name:        "positive test #1",
			shortURL:    "shortURL",
			originalURL: "https://practicum.yandex.ru/",
			userID:      "123",
			want: want{
				code:         http.StatusTemporaryRedirect,
				responseBody: "",
				location:     "https://practicum.yandex.ru/",
			},
		},
		{
			name: "negative test #1",
			want: want{
				code:         http.StatusNotFound,
				responseBody: settings.ErrOriginalURLNotFound.Error(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.shortURL) > 0 {
				repo.SaveShortURL(ctx, tt.shortURL, tt.originalURL, tt.userID)
			}

			request := httptest.NewRequest(http.MethodGet, "/"+tt.shortURL, nil).
				WithContext(context.WithValue(context.Background(), middleware.UserIDContextKey{}, tt.userID))
			w := httptest.NewRecorder()
			service := service.NewService(repo, request.Host)
			handler := NewHandler(service)
			handler.GetOriginalURL()(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)

			assert.Equal(t, tt.want.location, res.Header.Get("Location"))

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.want.responseBody, strings.TrimSuffix(string(resBody), "\n"))
		})
	}
}

func TestGetShortURLJSON(t *testing.T) {
	ctx := context.Background()
	repo := storage.NewLocalCahce()
	type input struct {
		URL string `json:"url"`
	}
	type output struct {
		Result string `json:"result"`
	}
	type want struct {
		code              int
		originalURLFromDB string
	}
	tests := []struct {
		name              string
		originalURLStruct input
		want              want
	}{
		{
			name:              "positive test #1",
			originalURLStruct: input{URL: "https://practicum.yandex.ru/"},
			want: want{
				code:              http.StatusCreated,
				originalURLFromDB: "https://practicum.yandex.ru/",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			body := httptest.NewRecorder().Body
			originalURLJSON, _ := json.Marshal(&tt.originalURLStruct)
			body.Write(originalURLJSON)
			request := httptest.NewRequest(http.MethodPost, "/", body).
				WithContext(context.WithValue(context.Background(), middleware.UserIDContextKey{}, "123"))

			w := httptest.NewRecorder()
			service := service.NewService(repo, request.Host)
			handler := NewHandler(service)
			handler.GetShortURLJSON()(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			var output output
			err = json.Unmarshal(resBody, &output)
			require.NoError(t, err)

			shortURL := output.Result[len(output.Result)-settings.ShortURLlen:]
			originalURLFromDB, err := repo.GetOriginalURL(ctx, shortURL)
			require.NoError(t, err)
			assert.Equal(t, tt.want.originalURLFromDB, originalURLFromDB)
		})
	}
}

func TestMarkRecordsForDeletion(t *testing.T) {
	ctx := context.Background()
	repo := storage.NewLocalCahce()
	type want struct {
		code         int
		responseBody string
	}
	tests := []struct {
		name        string
		shortURL    string
		originalURL string
		userID      string
		want        want
	}{
		{
			name:        "positive test #1",
			shortURL:    "shortURL",
			originalURL: "https://practicum.yandex.ru/",
			userID:      "123",
			want: want{
				code:         http.StatusAccepted,
				responseBody: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.shortURL) > 0 {
				repo.SaveShortURL(ctx, tt.shortURL, tt.originalURL, tt.userID)
			}
			var s []string
			s = append(s, tt.shortURL)
			data, _ := json.Marshal(s)
			body := httptest.NewRecorder().Body
			body.Write([]byte(data))
			request := httptest.NewRequest(http.MethodDelete, "/"+"api/user/urls", body).
				WithContext(context.WithValue(context.Background(), middleware.UserIDContextKey{}, tt.userID))
			w := httptest.NewRecorder()
			service := service.NewService(repo, request.Host)
			go service.HandleRecords()
			handler := NewHandler(service)
			handler.MarkRecordsForDeletion()(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.want.responseBody, strings.TrimSuffix(string(resBody), "\n"))
		})
	}
}

// func TestHandler_MarkRecordsForDeletion(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		h    *Handler
// 		want http.HandlerFunc
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := tt.h.MarkRecordsForDeletion(); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("Handler.MarkRecordsForDeletion() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
