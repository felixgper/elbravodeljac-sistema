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
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	defer s.Close(ctx)
	if err = s.Init(ctx); err != nil {
		log.Fatal("No se pudo inicializar/verificar el esquema MongoDB de desarrollo")
	}
	log.Print("MongoDB técnico de desarrollo inicializado y comprobado")
}
