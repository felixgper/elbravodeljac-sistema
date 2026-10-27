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

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/application"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/config"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/httpapi"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/mongodb"
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
	var technical application.Technical
	if uri := os.Getenv("MONGODB_URI"); uri != "" {
		name := os.Getenv("MONGODB_DATABASE")
		mongoStore, openErr := mongodb.Open(uri, name)
		if openErr != nil {
			log.Fatal(openErr)
		}
		technical = mongoStore
		defer func() {
			closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = mongoStore.Close(closeCtx)
		}()
	}
	frontendDir := os.Getenv("FRONTEND_DIR")
	if frontendDir == "" {
		frontendDir = "../dist"
	}
	handler, err := httpapi.New(&postgres.Store{Pool: pool}, httpapi.Options{FrontendDir: frontendDir, Origin: cfg.Origin, Key: cfg.Key, SecureCookies: cfg.SecureCookies, Technical: technical})
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
	log.Printf("Aplicación fases A–E: %s; MongoDB configurado: %t", cfg.Addr, technical != nil)
	if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Print("Servidor HTTP detenido por error")
	}
}
