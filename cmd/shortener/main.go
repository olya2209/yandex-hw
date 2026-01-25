package main

import (
	"log"

	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
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
		log.Fatal(err)
	}
	defer logger.Sync()
	sugar = *logger.Sugar()

	// dataBase
	db, err := sql.Open("pgx", cfg.Opts.DbAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	s, err := server.NewServer(cfg, sugar, db)
	if err != nil {
		sugar.Fatalln(err)
	}
	s.Run()
}
