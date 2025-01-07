package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/nasik90/url-shortener/internal/app/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetShortURL(t *testing.T) {
	cache := make(map[string]string)
	localCache := storage.LocalCache{CahceMap: cache}
	var mutex sync.Mutex
	type want struct {
		code              int
		originalURLFromDB string
	}
	tests := []struct {
		name        string
		originalURL string
		want        want
	}{
		{
			name:        "positive test #1",
			originalURL: "https://practicum.yandex.ru/",
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
			request := httptest.NewRequest(http.MethodPost, "/", body)
			w := httptest.NewRecorder()
			GetShortURL(&localCache, &mutex)(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)

			originalURLFromDB, err := localCache.GetOriginalURL(string(resBody))
			require.NoError(t, err)
			assert.Equal(t, tt.want.originalURLFromDB, originalURLFromDB)
		})
	}
}

func TestGetOriginalURL(t *testing.T) {
	cache := make(map[string]string)
	localCache := storage.LocalCache{CahceMap: cache}
	type want struct {
		code     int
		response string
	}
	tests := []struct {
		name        string
		shortURL    string
		originalURL string
		want        want
	}{
		{
			name:        "positive test #1",
			shortURL:    "shortURL",
			originalURL: "https://practicum.yandex.ru/",
			want: want{
				code:     http.StatusTemporaryRedirect,
				response: "https://practicum.yandex.ru/",
			},
		},
		{
			name: "negative test #1",
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.shortURL) > 0 {
				localCache.SaveShortURL(tt.shortURL, tt.originalURL)
			}

			request := httptest.NewRequest(http.MethodGet, "/"+tt.shortURL, nil)
			request.Pattern = "/{id}"
			w := httptest.NewRecorder()
			GetOriginalURL(&localCache)(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.response, res.Header.Get("Location"))
		})
	}
}
