package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	endpoint := "http://localhost:8080/W2h2M06Q"
	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		// Возвращаем ошибку, чтобы отключить перенаправление
		return http.ErrUseLastResponse
	}}
	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
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
	responseBody := string(body)
	fmt.Println(responseBody)
}
