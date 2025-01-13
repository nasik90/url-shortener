package settings

import (
	"flag"
	"os"
)

const (
	ShortURLlen            = 8
	OriginalURLNotFoundErr = "original URL not found"
)

type Options struct {
	A string
	B string
}

func ParseFlags(o *Options) {
	flag.StringVar(&o.A, "a", ":8080", "address and port to run server")
	flag.StringVar(&o.B, "b", "http://localhost:8080", "base address for short URL")
	flag.Parse()

	if serverAddress := os.Getenv("SERVER_ADDRESS"); serverAddress != "" {
		o.A = serverAddress
	}
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		o.B = baseURL
	}
}
