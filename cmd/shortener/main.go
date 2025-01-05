package main

import (
	"sync"

	"github.com/nasik90/url-shortener/internal/app/server"
	"github.com/nasik90/url-shortener/internal/app/storage"
)

func main() {

	cache := make(map[string]string)
	LocalCache := storage.LocalCache{CahceMap: cache}
	var mutex sync.Mutex
	server.RunServer(&LocalCache, &mutex)

}
