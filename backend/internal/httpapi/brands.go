package httpapi

import (
	"context"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"net/http"
)

type brandStore interface {
	Brands(context.Context) ([]string, error)
}

func (s *server) brands(w http.ResponseWriter, r *http.Request, u domain.User) {
	store, ok := s.store.(brandStore)
	if !ok {
		writeError(w, 503, "UNAVAILABLE", "No se pudieron consultar las marcas.")
		return
	}
	brands, err := store.Brands(r.Context())
	s.result(w, 200, brands, err)
}
