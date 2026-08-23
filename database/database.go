package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect() (*pgxpool.Pool, error) {
	connString := os.Getenv("DATABASE_URL")

	if connString == "" {
		return nil, fmt.Errorf("DATABASE_URL não definida no .env")
	}

	pool, err := pgxpool.New(context.Background(), connString)

	if err != nil {
		return nil, fmt.Errorf("Erro ao criar pool de conexões: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("Erro ao conectar no banco: %w", err)
	}

	return pool, nil
}
