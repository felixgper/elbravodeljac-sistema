// Package httpapi contiene el transporte HTTP. No accede directamente a bases.
package httpapi

import (
	"encoding/json"
	"net/http"
)

func New() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "mode": "architecture"})
	})
	// Solo certifica el proceso; no simula disponibilidad de PostgreSQL/MongoDB.
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusServiceUnavailable, "NOT_READY", "Adaptadores de persistencia pendientes.")
	})
	for _, route := range Routes {
		mux.HandleFunc(route.Method+" "+route.Path, func(w http.ResponseWriter, r *http.Request) {
			writeError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Contrato reservado; API de negocio pendiente.")
		})
	}
	return mux
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]string{"code": code, "message": message}})
}
