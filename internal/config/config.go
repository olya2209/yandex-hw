package config

import (
	"flag"
	"fmt"
	"net/url"
	"strings"

	"github.com/caarlos0/env/v6"
)

const defaultURL = "localhost:8080"

type Config struct {
	Opts *Options
}

type Options struct {
	Addr    string `env:"SERVER_ADDRESS"`
	BaseURL string `env:"BASE_URL"`
}

func newOpts() (*Options, error) {
	var (
		opt           Options
		addr, baseURL *string
	)

	err := env.Parse(&opt)
	if err != nil {
		return nil, err
	}

	if opt.Addr == "" {
		addr = flag.String("a", defaultURL, "server host")
	} else {
		addr = &opt.Addr
	}

	if _, err = url.Parse("https://" + *addr); err != nil {
		return nil, fmt.Errorf("incorrect parametr `-a` %s", *addr)
	}

	if opt.BaseURL == "" {
		baseURL = flag.String("b", defaultURL, "value before short URL")
		flag.Parse()
	} else {
		baseURL = &opt.BaseURL
	}

	if _, err = url.Parse("https://" + *baseURL); err != nil {
		return nil, fmt.Errorf("incorrect parametr `-b` %s", *baseURL)
	}

	if !strings.HasPrefix(*baseURL, "http://") && !strings.HasPrefix(*baseURL, "https://") {
		*baseURL = "http://" + *baseURL
	}
	*baseURL = strings.TrimSuffix(*baseURL, "/")

	return &Options{
		Addr:    *addr,
		BaseURL: *baseURL,
	}, nil
}

func NewConfig() (*Config, error) {
	opts, err := newOpts()
	if err != nil {
		return nil, err
	}
	return &Config{
		Opts: opts,
	}, nil
}
