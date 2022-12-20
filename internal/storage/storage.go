package storage

import (
	"context"
	"log"

	"github.com/T-V-N/gopherstore/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	Conn *pgxpool.Pool
	cfg  config.Config
}

func InitStorage(cfg config.Config) (*Storage, error) {
	m, err := migrate.New(
		"file://"+cfg.MigrationsPath,
		cfg.DatabaseURI)

	if err != nil {
		return nil, err
	}

	defer m.Close()

	err = m.Up()

	if err != nil {
		if err != migrate.ErrNoChange {
			return nil, err
		}
	}

	conn, err := pgxpool.New(context.Background(), cfg.DatabaseURI)

	if err != nil {
		log.Printf("Unable to connect to database: %v\n", err.Error())
		return nil, err
	}

	return &Storage{conn, cfg}, nil
}
