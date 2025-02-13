package main

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
	"github.com/nasik90/url-shortener/internal/app/server"
	"github.com/nasik90/url-shortener/internal/app/storage/pg"
)

func main() {
	options := new(settings.Options)
	settings.ParseFlags(options)

	// storage, err := storage.NewFileStorage(options.FilePath)
	// if err != nil {
	// 	panic(err)
	// }
	// defer func() {
	// 	err := storage.DestroyFileStorage()
	// 	if err != nil {
	// 		logger.Log.Error("destroy file storage", zap.String("info", "error to destroy file storage"), zap.String("error", err.Error()))
	// 	}
	// }()
	conn, err := sql.Open("pgx", options.DatabaseDSN)
	if err != nil {
		panic(err)
	}
	storage, err := pg.NewStore(conn)
	if err != nil {
		panic(err)
	}
	// storage := pg.NewStore(options.DatabaseDSN)
	err = server.RunServer(storage, options)
	if err != nil {
		panic(err)
	}
}
