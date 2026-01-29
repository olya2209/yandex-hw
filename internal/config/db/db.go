package db

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/olya2209/yandex-hw/internal/config"
	"go.uber.org/zap"
)

const dbDriver = "pgx"

func InitPostgresDB(cfg *config.Config, logger zap.SugaredLogger) (*sql.DB, error) {
	u, err := url.Parse(cfg.Opts.DbAddr)
	if err != nil {
		return nil, err
	}

	host := u.Host
	username := u.User.Username()
	password, _ := u.User.Password()
	dbname := strings.TrimPrefix(u.Path, "/")
	queryParams := u.Query()

	sslmode := queryParams.Get("sslmode")
	if sslmode == "" {
		sslmode = "disable"
	}

	addr := strings.Split(host, ":")
	if len(addr) != 2 {
		addr = []string{"postgres", "5432"}
	}

	dataSourceName := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		addr[0], addr[1], username, password, dbname, sslmode,
	)

	db, err := sql.Open(dbDriver, dataSourceName)
	if err != nil {
		log.Fatal(err)
	}

	logger.Infow(
		"InitDB",
		"addrDB", cfg.Opts.DbAddr,
		"pathDB", cfg.Opts.PathDB,
		"host", host,
		"dbname", dbname,
		"sslmode", sslmode,
	)

	return db, nil
}
