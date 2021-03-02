package database

import (
	"context"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

func GetDB() (*pgxpool.Pool, error) {
	datasource := os.ExpandEnv("host=${POSTGRES_HOST} port=${POSTGRES_PORT} user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${DATABASE_NAME} sslmode=disable")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return pgxpool.Connect(ctx, datasource)
}
