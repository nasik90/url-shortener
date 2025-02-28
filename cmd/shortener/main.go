package main

import (
	"database/sql"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nasik90/url-shortener/cmd/shortener/settings"
	"github.com/nasik90/url-shortener/internal/app/logger"
	"github.com/nasik90/url-shortener/internal/app/server"
	"github.com/nasik90/url-shortener/internal/app/service"
	"github.com/nasik90/url-shortener/internal/app/storage"
	"github.com/nasik90/url-shortener/internal/app/storage/pg"
	"go.uber.org/zap"
)

func main() {
	options := new(settings.Options)
	settings.ParseFlags(options)
	var (
		store service.Repository
		err   error
	)
	if options.DatabaseDSN != "" {
		conn, err := sql.Open("pgx", options.DatabaseDSN)
		if err != nil {
			logger.Log.Fatal("open pgx conn", zap.String("DatabaseDSN", options.DatabaseDSN), zap.String("error", err.Error()))
		}
		store, err = pg.NewStore(conn)
		if err != nil {
			logger.Log.Fatal("create pg store", zap.String("DatabaseDSN", options.DatabaseDSN), zap.String("error", err.Error()))
		}

	} else {
		store, err = storage.NewFileStorage(options.FilePath)
		if err != nil {
			logger.Log.Fatal("create file store", zap.String("FilePath", options.FilePath), zap.String("error", err.Error()))
		}
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		logger.Log.Info("closing the server")
		if err = store.Close(); err != nil {
			logger.Log.Error("close storage", zap.String("info", "error to close storage"), zap.String("error", err.Error()))
		}
		os.Exit(0)
	}()

	err = server.RunServer(store, options)
	if err != nil {
		logger.Log.Fatal("run server", zap.String("error", err.Error()))
	}
}
