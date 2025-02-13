package main

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nasik90/url-shortener/cmd/shortener/settings"
	"github.com/nasik90/url-shortener/internal/app/server"
	"github.com/nasik90/url-shortener/internal/app/service"
	"github.com/nasik90/url-shortener/internal/app/storage"
	"github.com/nasik90/url-shortener/internal/app/storage/pg"
)

func main() {
	options := new(settings.Options)
	settings.ParseFlags(options)
	var (
		store service.Repositories
		err   error
	)
	if options.DatabaseDSN != "" {
		conn, err := sql.Open("pgx", options.DatabaseDSN)
		if err != nil {
			panic(err)
		}
		store, err = pg.NewStore(conn)
		if err != nil {
			panic(err)
		}

	} else {
		store, err = storage.NewFileStorage(options.FilePath)
		if err != nil {
			panic(err)
		}
		// defer func() {
		// 	err = store.DestroyFileStorage()
		// 	if err != nil {
		// 		logger.Log.Error("destroy file storage", zap.String("info", "error to destroy file storage"), zap.String("error", err.Error()))
		// 	}
		// }()
	}
	// storage := pg.NewStore(options.DatabaseDSN)
	err = server.RunServer(store, options)
	if err != nil {
		panic(err)
	}
}
