package main

import (
	"log"

	"github.com/olya2209/yandex-hw/internal/config"
	"github.com/olya2209/yandex-hw/internal/server"
	"go.uber.org/zap"
)

var sugar zap.SugaredLogger

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalln(err)
	}
	// logging.
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		log.Fatal(err)
	}
	defer logger.Sync()
	sugar = *logger.Sugar()

	s, err := server.NewServer(cfg, sugar)
	if err != nil {
		sugar.Fatalln(err)
	}
	s.Run()
}
