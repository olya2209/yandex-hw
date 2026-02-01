package main

import (
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/olya2209/yandex-hw/internal/config"
	"github.com/olya2209/yandex-hw/internal/config/db"
	"github.com/olya2209/yandex-hw/internal/server"
	"github.com/olya2209/yandex-hw/migrations"
	"go.uber.org/zap"
)

var sugar zap.SugaredLogger

func main() {
	// logging
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()
	sugar = *logger.Sugar()

	// config: env and flags
	cfg, err := config.NewConfig(sugar)
	if err != nil {
		log.Fatalln(err)
	}

	// dataBase
	pgdb, _ := db.InitPostgresDB(cfg, sugar)
	defer pgdb.Close()

	sugar.Info("Running migrations...")
	err = migrations.Up(pgdb)
	if err != nil {
		sugar.Warn(err)
	}
	defer func() {
		migrations.Down(pgdb)
		sugar.Info("Migrations down")
	}()
	sugar.Info("Migrations applied successfully")

	s, err := server.NewServer(cfg, sugar, pgdb)
	if err != nil {
		sugar.Fatalln(err)
	}
	s.Run()
}
