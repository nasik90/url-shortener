// Пакет settings содержит настройки сервиса.
package settings

import (
	"errors"
	"flag"
	"os"
)

// Настройки короткого URL.
const (
	// ShortURLlen - длина короткого URL.
	ShortURLlen = 8
	// TemplateForRand - допустимые символы для формирования короткого URL.
	TemplateForRand = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
)

// Переменные - ошибки.
var (
	// ErrOriginalURLNotFound - ошибка - оригинальный URL не найден.
	ErrOriginalURLNotFound = errors.New("original URL not found")
	// ErrOriginalURLNotUnique - ошибка - оригинальный URL не уникальный.
	ErrOriginalURLNotUnique = errors.New("original URL is not unique")
	// ErrShortURLNotUnique - ошибка - короткий URL не найден.
	ErrShortURLNotUnique = errors.New("short URL is not unique")
)

// Options - структура для хранения настроек сервиса.
type Options struct {
	ServerAddress      string
	BaseURL            string
	LogLevel           string
	FilePath           string
	DatabaseDSN        string
	EnablePprofServ    bool
	PprofServerAddress string
}

// Record - структура для хранения короткого URL - UserID.
type Record struct {
	ShortURL string
	UserID   string
}

// ParseFlags - парсит флаги командной строки или переменные окружения.
// Результат сохраняет в структуру Options.
func ParseFlags(o *Options) {
	flag.StringVar(&o.ServerAddress, "a", ":8080", "address and port to run server")
	flag.StringVar(&o.BaseURL, "b", "http://localhost:8080", "base address for short URL")
	flag.StringVar(&o.LogLevel, "l", "debug", "log level")
	flag.StringVar(&o.FilePath, "f", "URLStorage.txt", "file storage path")
	//flag.StringVar(&o.DatabaseDSN, "d", "host=localhost user=postgres password=xxxx dbname=URLShortener sslmode=disable", "database connection string")
	flag.StringVar(&o.DatabaseDSN, "d", "", "database connection string")
	flag.BoolVar(&o.EnablePprofServ, "p", true, "enable pprof server")
	flag.StringVar(&o.PprofServerAddress, "pa", ":8181", "address and port to run pprof server")
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
