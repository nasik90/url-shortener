package settings

import "flag"

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
}
