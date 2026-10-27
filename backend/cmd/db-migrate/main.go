package main

import (
	"context"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/config"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
	"time"
)

func main() {
	dsn := os.Getenv("POSTGRES_DSN")
	if err := config.ValidateDSN(dsn); err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal("No se pudo preparar la migración aislada")
	}
	defer pool.Close()
	if err = (&postgres.Store{Pool: pool}).Migrate(ctx); err != nil {
		log.Fatal("Migración fallida; no se publican credenciales: revisar PostgreSQL de desarrollo")
	}
	log.Print("Migración funcional aplicada sin eliminar tablas, FK ni datos")
}
