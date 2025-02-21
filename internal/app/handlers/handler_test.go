package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
	"github.com/nasik90/url-shortener/internal/app/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetShortURL(t *testing.T) {
	ctx := context.Background()
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

			storage, err := createFileStorage("TestGetShortURL*.txt")
			require.NoError(t, err)

			body := httptest.NewRecorder().Body
			body.Write([]byte(tt.originalURL))
			request := httptest.NewRequest(http.MethodPost, "/", body)
			w := httptest.NewRecorder()
			GetShortURL(storage, request.Host)(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			resBodyString := string(resBody)
			shortURL := string(resBody)[len(resBodyString)-settings.ShortURLlen:]
			originalURLFromDB, err := storage.GetOriginalURL(ctx, shortURL)
			require.NoError(t, err)
			assert.Equal(t, tt.want.originalURLFromDB, originalURLFromDB)
		})
	}
}

func TestGetOriginalURL(t *testing.T) {
	ctx := context.Background()
	cache := make(map[string]string)
	localCache := storage.LocalCache{CahceMap: cache}
	type want struct {
		code         int
		responseBody string
		location     string
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
				localCache.SaveShortURL(ctx, tt.shortURL, tt.originalURL)
			}

			request := httptest.NewRequest(http.MethodGet, "/"+tt.shortURL, nil)
			w := httptest.NewRecorder()
			GetOriginalURL(&localCache)(w, request)

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

			storage, err := createFileStorage("TestGetShortURLJSON*.txt")
			require.NoError(t, err)

			body := httptest.NewRecorder().Body
			originalURLJSON, _ := json.Marshal(&tt.originalURLStruct)
			body.Write(originalURLJSON)
			request := httptest.NewRequest(http.MethodPost, "/", body)
			w := httptest.NewRecorder()
			GetShortURLJSON(storage, request.Host)(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			var output output
			err = json.Unmarshal(resBody, &output)
			require.NoError(t, err)

			shortURL := output.Result[len(output.Result)-settings.ShortURLlen:]
			originalURLFromDB, err := storage.GetOriginalURL(ctx, shortURL)
			require.NoError(t, err)
			assert.Equal(t, tt.want.originalURLFromDB, originalURLFromDB)
		})
	}
}

func createFileStorage(fileName string) (*storage.FileStorage, error) {

	file, err := os.CreateTemp("", fileName)
	if err != nil {
		return nil, err
	}
	return storage.NewFileStorage(file.Name())
}
