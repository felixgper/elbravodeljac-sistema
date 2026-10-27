package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/config"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/httpapi"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	pc, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		log.Fatal("Configuración PostgreSQL inválida")
	}
	pc.MaxConns = 10
	pc.ConnConfig.RuntimeParams["timezone"] = "UTC"
	pc.ConnConfig.ConnectTimeout = 5 * time.Second
	pool, err := pgxpool.NewWithConfig(ctx, pc)
	if err != nil {
		log.Fatal("No se pudo crear el pool")
	}
	defer pool.Close()
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	err = pool.Ping(pingCtx)
	cancel()
	if err != nil {
		log.Fatal("PostgreSQL no disponible; comprueba el entorno de pruebas")
	}
	handler, err := httpapi.New(&postgres.Store{Pool: pool}, httpapi.Options{Origin: cfg.Origin, Key: cfg.Key, SecureCookies: cfg.SecureCookies})
	if err != nil {
		log.Fatal("No se pudo preparar HTTP")
	}
	server := &http.Server{
		Addr: cfg.Addr, Handler: handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second, WriteTimeout: 15 * time.Second,
		IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	log.Printf("API fase A: %s", cfg.Addr)
	if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Print("Servidor HTTP detenido por error")
	}
}
