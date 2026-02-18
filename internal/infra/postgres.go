package infra

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgres(databaseURL string) *pgxpool.Pool {
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("postgres: unable to create pool: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("postgres: ping failed: %v", err)
	}
	log.Println("postgres: connected")
	return pool
}
