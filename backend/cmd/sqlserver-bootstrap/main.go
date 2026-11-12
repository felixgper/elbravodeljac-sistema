// Explicit provisioning for a fresh, isolated SQL Server database.
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/sqlserver"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	store, err := sqlserver.Open(os.Getenv("SQLSERVER_DSN"))
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()
	if err := store.ProvisionInstallation(ctx, os.Getenv("SQLSERVER_BOOTSTRAP_CONFIRM"),
		os.Getenv("DEV_ADMIN_USERNAME"), os.Getenv("DEV_ADMIN_PASSWORD"),
		os.Getenv("JAC_ADMIN_PASSWORD"), os.Getenv("JAC_MECHANIC_PASSWORD")); err != nil {
		log.Fatal("La provisión SQL Server falló: ", err)
	}
	log.Print("Catálogos, permisos y tres cuentas iniciales preparados.")
}
