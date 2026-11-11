// The command reads only from isolated copies. See database/sqlserver/MIGRACION_DATOS.md.
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/datamigration"
)

func main() {
	timeout := 30 * time.Minute
	if value := os.Getenv("MIGRATION_TIMEOUT"); value != "" {
		var err error
		timeout, err = time.ParseDuration(value)
		if err != nil || timeout <= 0 {
			log.Fatal("MIGRATION_TIMEOUT inválido")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	report, err := datamigration.Run(ctx, datamigration.Config{
		Source: os.Getenv("MIGRATION_POSTGRES_DSN"),
		Target: os.Getenv("MIGRATION_SQLSERVER_DSN"),
		Mongo:  os.Getenv("MIGRATION_MONGODB_URI"),
		Marker: os.Getenv("MIGRATION_RUN_TOKEN"),
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
		log.Fatal("Carga confirmada. No se pudo emitir el informe de conteos.")
	}
}
