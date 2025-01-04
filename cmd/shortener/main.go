package main

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/nasik90/url-shortener/internal/app/handlers"
)

func mainPage(res http.ResponseWriter, req *http.Request) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		//res.Write([]byte(err.Error()))
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	url := buf.String()
	randomString, err := handlers.RandomString(8)
	if err != nil {
		//res.Write([]byte(err.Error()))
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	shortUrl := req.Host + "/" + randomString
	urlCache := make(map[string]string)
	urlCache[shortUrl] = url
	res.Header().Set("content-type", "text/plain")
	res.Header().Set("Content-Length", strconv.Itoa(len(shortUrl)))
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(shortUrl))
}

func urlShort(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte("Привет!"))
}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", mainPage)
	mux.HandleFunc("/{id}", urlShort)
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}

}
