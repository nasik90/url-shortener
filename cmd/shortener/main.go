package main

import (
	"database/sql"
	"errors"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
	handler "github.com/nasik90/url-shortener/internal/app/handlers"
	"github.com/nasik90/url-shortener/internal/app/logger"
	"github.com/nasik90/url-shortener/internal/app/server"
	"github.com/nasik90/url-shortener/internal/app/service"
	"github.com/nasik90/url-shortener/internal/app/storage"
	"github.com/nasik90/url-shortener/internal/app/storage/pg"
)

func main() {

	options := new(settings.Options)
	settings.ParseFlags(options)

	if err := logger.Initialize(options.LogLevel); err != nil {
		panic(err)
	}

	if options.EnablePprofServ {
		go func() {
			if err := http.ListenAndServe(options.PprofServerAddress, nil); err != nil {
				logger.Log.Error("run pprof server", zap.Error(err))
			}
			logger.Log.Info("run pprof server", zap.String("addr", options.PprofServerAddress))
		}()
	}

	var (
		repo service.Repository
		err  error
	)
	if options.DatabaseDSN != "" {
		conn, err := sql.Open("pgx", options.DatabaseDSN)
		if err != nil {
			logger.Log.Fatal("open pgx conn", zap.String("DatabaseDSN", options.DatabaseDSN), zap.String("error", err.Error()))
		}
		repo, err = pg.NewStore(conn)
		if err != nil {
			logger.Log.Fatal("create pg repo", zap.String("DatabaseDSN", options.DatabaseDSN), zap.String("error", err.Error()))
		}

	} else if options.FilePath != "" {
		repo, err = storage.NewFileStorage(options.FilePath)
		if err != nil {
			logger.Log.Fatal("create file repo", zap.String("FilePath", options.FilePath), zap.String("error", err.Error()))
		}

	} else {
		repo = storage.NewLocalCahce()
	}

	service := service.NewService(repo, options.BaseURL)
	handler := handler.NewHandler(service)
	server := server.NewServer(handler, options.ServerAddress)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-sigs
		logger.Log.Info("closing the server")
		if err := server.StopServer(); err != nil {
			logger.Log.Error("stop http server", zap.String("error", err.Error()))
		}
		logger.Log.Info("closing the storage")
		if err := repo.Close(); err != nil {
			logger.Log.Error("close storage", zap.String("error", err.Error()))
		}
		logger.Log.Info("ready to exit")
	}()

	go service.HandleRecords()

	err = server.RunServer()
	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Fatal("run server", zap.String("error", err.Error()))
		}
	}

	wg.Wait()
	logger.Log.Info("closed gracefuly")

}
