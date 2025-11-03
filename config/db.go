package config

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool() (*pgxpool.Pool, error) {
	dsn := "postgres://leanq@localhost:5432/techup_db?sslmode=disable"

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to db: %w", err)
	}

	return pool, nil
}
