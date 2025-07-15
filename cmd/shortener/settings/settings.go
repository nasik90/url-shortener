// Пакет settings содержит настройки сервиса.
package settings

import (
	"encoding/json"
	"errors"
	"flag"
	"os"
	"strconv"
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

type methodToCheckTrustedNet struct{ GRPSMethod, APIMethod string }

// MethodsToCheckTrustedNet - массив с именами методов для проверки доверенной сети.
var MethodsToCheckTrustedNet = [1]methodToCheckTrustedNet{{GRPSMethod: "GetURLsStats", APIMethod: "/api/internal/stats"}}

// Options - структура для хранения настроек сервиса.
type Options struct {
	ServerAddress      string `json:"server_address"`
	BaseURL            string `json:"base_url"`
	LogLevel           string `json:"log_level"`
	FilePath           string `json:"file_storage_path"`
	DatabaseDSN        string `json:"database_dsn"`
	EnablePprofServ    bool   `json:"enable_pprof_server"`
	PprofServerAddress string `json:"pprof_server_address"`
	EnableHTTPS        bool   `json:"enable_https,omitempty"`
	Config             string
	TrustedSubnet      string `json:"trusted_subnet"`
	GRPCServerAddress  string `json:"grps_server_address"`
}

// Record - структура для хранения короткого URL - UserID.
type Record struct {
	ShortURL string
	UserID   string
}

// ParseFlags - парсит флаги командной строки или переменные окружения.
// Результат сохраняет в структуру Options.
func ParseFlags(o *Options) {
	//flag.StringVar(&o.Config, "c", "config.json", "config path")
	flag.StringVar(&o.Config, "c", "", "config path")
	if config := os.Getenv("CONFIG"); config != "" {
		o.Config = config
	}

	var config Options
	var err error
	if o.Config != "" {
		config, err = readConfig(o.Config)
		if err != nil {
			panic(err)
		}
	}

	fillDefaultOptions(o)
	overrideOptionsFromConfig(o, &config)
	overrideOptionsFromCmd(o)
	overrideOptionsFromEnv(o)

}

func fillDefaultOptions(o *Options) {
	o.ServerAddress = ":8080"
	o.BaseURL = "http://localhost:8080"
	o.LogLevel = "debug"
	o.FilePath = "URLStorage.txt"
	//o.DatabaseDSN = "host=localhost user=postgres password=xxxx dbname=URLShortener sslmode=disable"
	o.DatabaseDSN = ""
	o.EnablePprofServ = true
	o.PprofServerAddress = ":8181"
	o.EnableHTTPS = false
	o.TrustedSubnet = "192.168.0.1/24"
	o.GRPCServerAddress = ":3200"
}

func overrideOptionsFromConfig(o *Options, c *Options) {
	if c.ServerAddress != "" {
		o.ServerAddress = c.ServerAddress
	}
	if c.BaseURL != "" {
		o.BaseURL = c.BaseURL
	}
	if c.LogLevel != "" {
		o.LogLevel = c.LogLevel
	}
	if c.FilePath != "" {
		o.FilePath = c.FilePath
	}
	if c.DatabaseDSN != "" {
		o.DatabaseDSN = c.DatabaseDSN
	}
	o.EnablePprofServ = c.EnablePprofServ
	if c.PprofServerAddress != "" {
		o.PprofServerAddress = c.PprofServerAddress
	}
	o.EnableHTTPS = c.EnableHTTPS
	if c.TrustedSubnet != "" {
		o.TrustedSubnet = c.TrustedSubnet
	}
	if c.GRPCServerAddress != "" {
		o.GRPCServerAddress = c.GRPCServerAddress
	}
}

func readConfig(fname string) (Options, error) {
	var config Options

	f, err := os.Open(fname)
	if err != nil {
		return config, err
	}
	out := make([]byte, 1024)
	var n int
	if n, err = f.Read(out); err != nil {
		return config, err
	}
	data := out[:n]
	err = json.Unmarshal(data, &config)
	return config, err

}

func overrideOptionsFromCmd(o *Options) {
	flag.StringVar(&o.ServerAddress, "a", o.ServerAddress, "address and port to run server")
	flag.StringVar(&o.BaseURL, "b", o.BaseURL, "base address for short URL")
	flag.StringVar(&o.LogLevel, "l", o.LogLevel, "log level")
	flag.StringVar(&o.FilePath, "f", o.FilePath, "file storage path")
	//flag.StringVar(&o.DatabaseDSN, "d", o.DatabaseDSN, "database connection string")
	flag.StringVar(&o.DatabaseDSN, "d", o.DatabaseDSN, "database connection string")
	flag.BoolVar(&o.EnablePprofServ, "p", o.EnablePprofServ, "enable pprof server")
	flag.StringVar(&o.PprofServerAddress, "pa", o.PprofServerAddress, "address and port to run pprof server")
	flag.BoolVar(&o.EnableHTTPS, "s", o.EnableHTTPS, "enable HTPPS connection")
	flag.StringVar(&o.TrustedSubnet, "t", o.TrustedSubnet, "trusted subnet")
	flag.StringVar(&o.GRPCServerAddress, "ga", o.GRPCServerAddress, "address and port to run gRPC server")
	flag.Parse()
}

func overrideOptionsFromEnv(o *Options) {

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
	if enableHTTPS := os.Getenv("ENABLE_HTTPS"); enableHTTPS != "" {
		val, err := strconv.ParseBool(enableHTTPS)
		if err != nil {
			panic("error parsing env var ENABLE_HTTPS: " + err.Error())
		}
		o.EnableHTTPS = val
	}
	if trustedSubnet := os.Getenv("TRUSTED_SUBNET"); trustedSubnet != "" {
		o.TrustedSubnet = trustedSubnet
	}
	if gRPCServerAddress := os.Getenv("GRPC_SERVER_ADDRESS"); gRPCServerAddress != "" {
		o.GRPCServerAddress = gRPCServerAddress
	}
}
