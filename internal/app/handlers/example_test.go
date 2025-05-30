package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
)

func ExampleHandler_GetShortURL() {

	options := new(settings.Options)
	settings.ParseFlags(options)

	originalURL := "http://ya.ru"
	endpoint := options.BaseURL
	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		// Возвращаем ошибку, чтобы отключить перенаправление
		return http.ErrUseLastResponse
	}}
	// Запросим короткий URL
	request, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(originalURL))
	if err != nil {
		panic(err)
	}
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	fmt.Println(response.StatusCode)
	responseBody := string(body)

	// Запросим оригинальный URL для проверки, что вернется корректный URL
	request, err = http.NewRequest(http.MethodGet, responseBody, nil)
	if err != nil {
		panic(err)
	}
	response, err = client.Do(request)
	if err != nil {
		panic(err)
	}
	originalURL = response.Header.Get("Location")
	fmt.Println(response.StatusCode)
	fmt.Println(originalURL)

	// Output:
	// 201
	// 307
	// http://ya.ru
}

func ExampleHandler_GetOriginalURL() {

	options := new(settings.Options)
	settings.ParseFlags(options)

	originalURL := "http://habr.ru"
	endpoint := options.BaseURL
	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		// Возвращаем ошибку, чтобы отключить перенаправление
		return http.ErrUseLastResponse
	}}
	// Запросим короткий URL, чтобы в дальнейшем получить оригинальный URL
	request, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(originalURL))
	if err != nil {
		panic(err)
	}
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(response.StatusCode)
	responseBody := string(body)

	// Запросим оригинальный URL
	request, err = http.NewRequest(http.MethodGet, responseBody, nil)
	if err != nil {
		panic(err)
	}
	response, err = client.Do(request)
	if err != nil {
		panic(err)
	}
	originalURL = response.Header.Get("Location")
	fmt.Println(response.StatusCode)
	fmt.Println(originalURL)

	// Output:
	// 201
	// 307
	// http://habr.ru
}
