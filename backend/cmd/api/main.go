package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/httpapi"
)

func main() {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	server := &http.Server{
		Addr: addr, Handler: httpapi.New(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second, WriteTimeout: 15 * time.Second,
		IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20,
	}
	log.Printf("API: %s (arquitectura; sin conexiones a bases de datos)", addr)
	log.Fatal(server.ListenAndServe())
}
