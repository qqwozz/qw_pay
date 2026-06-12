package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qw_pay/internal/config"
)

// Pool — пул соединений с PostgreSQL, используемый всеми сервисами.
var Pool *pgxpool.Pool

// Connect устанавливает соединение с базой данных и проверяет доступность.
func Connect() {
	ctx := context.Background()
	var err error
	Pool, err = pgxpool.New(ctx, config.C.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	if err = Pool.Ping(ctx); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}
	log.Println("Connected to PostgreSQL")
}

// Close закрывает пул соединений при завершении работы приложения.
func Close() {
	Pool.Close()
}
