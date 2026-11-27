// Explicit provisioning for the satisfaction module. Never run at API startup.
package main

import (
	"context"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/mongodb"
	"log"
	"os"
	"time"
)

func main() {
	s, err := mongodb.Open(os.Getenv("MONGODB_URI"), os.Getenv("MONGODB_DATABASE"))
	if err != nil {
		log.Fatal("Configuración MongoDB no disponible")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	defer s.Close(ctx)
	if err = s.InitSurveys(ctx); err != nil {
		log.Fatal("No se pudo inicializar/verificar las encuestas; revisar permisos y esquema sin modificar colecciones existentes")
	}
	log.Print("Colección e índices de encuestas verificados")
}
