package main

import (
	"context"
	"flag"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/mongodb"
	"log"
	"os"
	"time"
)

func main() {
	repair := flag.Bool("repair", false, "Limpiar candidatos huérfanos antiguos; por defecto solo informa")
	flag.Parse()
	s, err := mongodb.Open(os.Getenv("MONGODB_URI"), os.Getenv("MONGODB_DATABASE"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	defer s.Close(ctx)
	n, err := s.Reconcile(ctx, !*repair)
	if err != nil {
		log.Fatal("No se pudo reconciliar MongoDB técnico")
	}
	log.Printf("Candidatos huérfanos: %d; reparación: %t", n, *repair)
}
