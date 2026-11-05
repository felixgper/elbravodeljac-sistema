package httpapi

import (
	"context"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"net/http"
)

type documentStore interface {
	DocumentTypes(context.Context) (any, error)
}

func (s *server) documentTypes(w http.ResponseWriter, r *http.Request, u domain.User) {
	if !u.Can("clientes.ver") && !u.Can("pagos.ver") {
		s.failure(w, domain.ErrForbidden)
		return
	}
	store, ok := s.store.(documentStore)
	if !ok {
		writeError(w, 503, "UNAVAILABLE", "Catálogo documental no disponible.")
		return
	}
	out, err := store.DocumentTypes(r.Context())
	s.result(w, 200, out, err)
}
