package main

import (
	"log"

	"github.com/olya2209/yandex-hw/internal/config"
	"github.com/olya2209/yandex-hw/internal/logger"
	"github.com/olya2209/yandex-hw/internal/server"
)

func main() {
	cfg, err := config.NewConfig()

	sugar, err := logger.NewLogger(cfg.Opts.Addr)
	if err != nil {
		log.Fatalln(err)
		panic(err)
	}

	s := server.NewServer(cfg, sugar)
	s.Run()
}
