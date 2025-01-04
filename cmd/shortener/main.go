package main

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/nasik90/url-shortener/internal/app/handlers"
)

const shortURLlen = 8

var urlCache = make(map[string]string)

func mainPage(res http.ResponseWriter, req *http.Request) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		//res.Write([]byte(err.Error()))
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	url := buf.String()
	randomString, err := handlers.RandomString(shortURLlen)
	if err != nil {
		//res.Write([]byte(err.Error()))
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	shortURL := req.Host + "/" + randomString
	//urlCache := make(map[string]string)
	urlCache[shortURL] = url
	res.Header().Set("content-type", "text/plain")
	res.Header().Set("Content-Length", strconv.Itoa(len(shortURL)))
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(shortURL))
}

func urlShort(res http.ResponseWriter, req *http.Request) {
	id := req.RequestURI
	if len(id) == shortURLlen+1 {
		id = id[1:]
	}
	shortURL := req.Host + "/" + id
	originalURL, ok := urlCache[shortURL]
	if !ok {
		http.Error(res, "nothing found", http.StatusBadRequest)
		return
	}
	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
	//res.Write([]byte(id))
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
