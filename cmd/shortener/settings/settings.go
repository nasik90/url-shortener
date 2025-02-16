package settings

import (
	"errors"
	"flag"
	"os"
)

const (
	ShortURLlen     = 8
	TemplateForRand = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
)

var (
	ErrOriginalURLNotFound  = errors.New("original URL not found")
	ErrOriginalURLNotUnique = errors.New("original URL is not unique")
)

type Options struct {
	ServerAddress string
	BaseURL       string
	LogLevel      string
	FilePath      string
	DatabaseDSN   string
}

func ParseFlags(o *Options) {
	flag.StringVar(&o.ServerAddress, "a", ":8080", "address and port to run server")
	flag.StringVar(&o.BaseURL, "b", "http://localhost:8080", "base address for short URL")
	flag.StringVar(&o.LogLevel, "l", "debug", "log level")
	flag.StringVar(&o.FilePath, "f", "URLStorage.txt", "file storage path")
	//flag.StringVar(&o.DatabaseDSN, "d", "host=localhost user=postgres password=xxxx dbname=URLShortener sslmode=disable", "database connection string")
	flag.StringVar(&o.DatabaseDSN, "d", "", "database connection string")
	flag.Parse()

	if serverAddress := os.Getenv("SERVER_ADDRESS"); serverAddress != "" {
		o.ServerAddress = serverAddress
	}
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		o.BaseURL = baseURL
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		o.LogLevel = envLogLevel
	}
	if filePath := os.Getenv("FILE_STORAGE_PATH"); filePath != "" {
		o.FilePath = filePath
	}
	if databaseDSN := os.Getenv("DATABASE_DSN"); databaseDSN != "" {
		o.DatabaseDSN = databaseDSN
	}
}
