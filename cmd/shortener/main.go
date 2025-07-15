package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
	"github.com/nasik90/url-shortener/internal/app/grpcapi"
	"github.com/nasik90/url-shortener/internal/app/grpcserver"
	handler "github.com/nasik90/url-shortener/internal/app/handlers"
	"github.com/nasik90/url-shortener/internal/app/logger"
	"github.com/nasik90/url-shortener/internal/app/server"
	"github.com/nasik90/url-shortener/internal/app/service"
	"github.com/nasik90/url-shortener/internal/app/storage"
	"github.com/nasik90/url-shortener/internal/app/storage/pg"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printFlags()

	options := parseOptions()

	if err := logger.Initialize(options.LogLevel); err != nil {
		panic(err)
	}

	startPprofIfEnabled(options)

	repo := initRepository(options)

	service := service.NewService(repo, options.BaseURL)
	handler := handler.NewHandler(service, options.TrustedSubnet)
	shortenerServer := grpcapi.NewShortenerServer(service)
	server := server.NewServer(handler, options.ServerAddress, options.EnableHTTPS, options.TrustedSubnet)
	grpcServer := grpcserver.NewGRPCServer(shortenerServer, options.GRPCServerAddress, options.TrustedSubnet)

	go service.HandleRecords()

	runServers(server, grpcServer, repo)
}

func printFlags() {

	buildInfoMap := make(map[string]string)
	buildInfoMap["Build version"] = buildVersion
	buildInfoMap["Build date"] = buildDate
	buildInfoMap["Build commit"] = buildCommit

	for key, value := range buildInfoMap {
		if value == "" {
			fmt.Printf("%s: N/A", key)
		} else {
			fmt.Printf("%s: %s\n", key, value)
		}

	}
}

func parseOptions() *settings.Options {
	options := new(settings.Options)
	settings.ParseFlags(options)
	return options
}

func startPprofIfEnabled(options *settings.Options) {
	if !options.EnablePprofServ {
		return
	}
	go func() {
		if err := http.ListenAndServe(options.PprofServerAddress, nil); err != nil {
			logger.Log.Error("run pprof server", zap.Error(err))
		}
		logger.Log.Info("run pprof server", zap.String("addr", options.PprofServerAddress))
	}()
}

func initRepository(options *settings.Options) service.Repository {
	var (
		repo service.Repository
		err  error
	)

	if options.DatabaseDSN != "" {
		conn, err := sql.Open("pgx", options.DatabaseDSN)
		if err != nil {
			logger.Log.Fatal("open pgx conn", zap.String("DatabaseDSN", options.DatabaseDSN), zap.Error(err))
		}
		repo, err = pg.NewStore(conn)
		if err != nil {
			logger.Log.Fatal("create pg repo", zap.String("DatabaseDSN", options.DatabaseDSN), zap.Error(err))
		}
	} else if options.FilePath != "" {
		repo, err = storage.NewFileStorage(options.FilePath)
		if err != nil {
			logger.Log.Fatal("create file repo", zap.String("FilePath", options.FilePath), zap.Error(err))
		}
	} else {
		repo = storage.NewLocalCahce()
	}

	return repo
}

func runServers(server *server.Server, grpcServer *grpcserver.GRPCServer, repo service.Repository) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.RunServer(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Fatal("run server", zap.Error(err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := grpcServer.RunServer(); err != nil {
			logger.Log.Fatal("run grpc server", zap.Error(err))
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-sigs
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		logger.Log.Info("closing http server")
		if err := server.StopServer(ctx); err != nil {
			logger.Log.Error("stop http server", zap.Error(err))
		}

		logger.Log.Info("closing grpc server")
		grpcServer.StopServer()

		logger.Log.Info("closing the storage")
		if err := repo.Close(); err != nil {
			logger.Log.Error("close storage", zap.Error(err))
		}

		logger.Log.Info("ready to exit")
	}()

	wg.Wait()
	logger.Log.Info("closed gracefully")
}
