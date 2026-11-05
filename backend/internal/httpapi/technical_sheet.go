package httpapi

import (
	"context"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/technicalsheet"
	"net/http"
)

type sheetStore interface {
	Inspection(context.Context, string, domain.User) (any, error)
	SaveInspection(context.Context, string, technicalsheet.Control, domain.User) (any, error)
	SheetPDF(context.Context, string, domain.User) ([]byte, error)
}

func (s *server) inspection(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	store, ok := s.store.(sheetStore)
	if !ok {
		writeError(w, 503, "UNAVAILABLE", "Control técnico no disponible.")
		return
	}
	var out any
	var err error
	if r.Method == "PUT" {
		var v technicalsheet.Control
		if !decodeLimit(w, r, &v, 1<<20) {
			return
		}
		out, err = store.SaveInspection(r.Context(), id, v, u)
	} else {
		out, err = store.Inspection(r.Context(), id, u)
	}
	if err != nil {
		s.failure(w, err)
		return
	}
	respond(w, 200, map[string]any{"data": map[string]any{"record": out, "catalogue": technicalsheet.Catalog}})
}
func (s *server) sheet(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	store, ok := s.store.(sheetStore)
	if !ok {
		writeError(w, 503, "UNAVAILABLE", "Ficha no disponible.")
		return
	}
	pdf, err := store.SheetPDF(r.Context(), id, u)
	if err != nil {
		s.failure(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="ficha-OT-`+id+`.pdf"`)
	w.WriteHeader(200)
	w.Write(pdf)
}
