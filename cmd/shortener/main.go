package main

import (
	"github.com/nasik90/url-shortener/cmd/shortener/settings"
	"github.com/nasik90/url-shortener/internal/app/logger"
	"github.com/nasik90/url-shortener/internal/app/server"
	"github.com/nasik90/url-shortener/internal/app/storage"
	"go.uber.org/zap"
)

func main() {
	options := new(settings.Options)
	settings.ParseFlags(options)

	// cache := make(map[string]string)
	// localCache := &storage.LocalCache{CahceMap: cache}
	// server.RunServer(localCache, options)
	storage, err := storage.NewFileStorage(options.FilePath)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := storage.DestroyFileStorage()
		if err != nil {
			logger.Log.Error("destroy file storage", zap.String("info", "error to destroy file storage"), zap.String("error", err.Error()))
		}
	}()
	server.RunServer(storage, options)
}
