package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func ExampleHandler_GetShortURL() {

	shortURL := "ya.ru"
	endpoint := "http://localhost:8080/"
	client := &http.Client{}
	request, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(shortURL))
	if err != nil {
		panic(err)
	}
	// в заголовках запроса указываем кодировку
	// request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))

}

func ExampleHandler_GetOriginalURL() {

	endpoint := "http://localhost:8080/qwErty12"
	client := &http.Client{}
	// Запросим оригинальный URL
	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		panic(err)
	}
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	originalURL := response.Header.Get("Location")
	fmt.Println(response.StatusCode)
	fmt.Println(originalURL)
}
