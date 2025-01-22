package main

import (
	"github.com/nasik90/url-shortener/cmd/shortener/settings"
	"github.com/nasik90/url-shortener/internal/app/server"
	"github.com/nasik90/url-shortener/internal/app/storage"
)

func main() {
	options := new(settings.Options)
	settings.ParseFlags(options)

	cache := make(map[string]string)
	localCache := &storage.LocalCache{CahceMap: cache}
	server.RunServer(localCache, options)

}
