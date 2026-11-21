package db

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Pool *pgxpool.Pool
}

func Connect(databaseURL string) *Database {
	// Connection pool

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Не удалось создать пул подключений: %v\n", err)
	}

	// Check real connection with timeout 5 sec
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("База данных недоступна: %v\n", err)
	}

	log.Println("🚀 Успешно подключено к PostgreSQL!")

	return &Database{Pool: pool}
}

func (d *Database) Close() {
	d.Pool.Close()
}
