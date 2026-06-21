package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qw_pay/internal/config"
	"github.com/qw_pay/internal/logger"
)

var Pool *pgxpool.Pool

func Connect() {
	ctx := context.Background()
	var err error
	Pool, err = pgxpool.New(ctx, config.C.DatabaseURL)
	if err != nil {
		logger.Error("unable to connect to database", "error", err)
		panic(fmt.Sprintf("unable to connect to database: %v", err))
	}
	if err = Pool.Ping(ctx); err != nil {
		logger.Error("unable to ping database", "error", err)
		panic(fmt.Sprintf("unable to ping database: %v", err))
	}
	logger.Info("connected to PostgreSQL")
}

func ConnectWithDSN(dsn string) (*pgxpool.Pool, error) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

func Close() {
	if Pool != nil {
		Pool.Close()
	}
}
