package httpapi

import (
	"context"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"net/http"
)

type quotePDFStore interface {
	QuotePDF(context.Context, string, domain.User) ([]byte, error)
}

func (s *server) quotePDF(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	store, ok := s.store.(quotePDFStore)
	if !ok {
		writeError(w, 503, "UNAVAILABLE", "Exportación no disponible.")
		return
	}
	pdf, err := store.QuotePDF(r.Context(), id, u)
	if err != nil {
		s.failure(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="cotizacion-`+id+`.pdf"`)
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(http.StatusOK)
	w.Write(pdf)
}
