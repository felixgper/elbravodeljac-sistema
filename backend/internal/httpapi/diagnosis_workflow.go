package httpapi

import (
	"context"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"net/http"
)

type diagnosisWorkflowStore interface {
	AssignDiagnosis(context.Context, string, domain.DiagnosisAssignment, domain.User) (any, error)
	SaveDiagnosis(context.Context, string, domain.DiagnosisInput, domain.User) (any, error)
	DiagnosisQueue(context.Context, bool, domain.Page, domain.User) (any, error)
}

func (s *server) diagnosisQueue(ready bool) func(http.ResponseWriter, *http.Request, domain.User) {
	return func(w http.ResponseWriter, r *http.Request, u domain.User) {
		store, ok := s.store.(diagnosisWorkflowStore)
		if !ok {
			writeError(w, 503, "UNAVAILABLE", "Flujo de diagnóstico no disponible.")
			return
		}
		p, err := page(r)
		if err != nil {
			s.failure(w, err)
			return
		}
		out, err := store.DiagnosisQueue(r.Context(), ready, p, u)
		if err != nil {
			s.failure(w, err)
			return
		}
		respond(w, 200, out)
	}
}
func (s *server) assignDiagnosis(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	store, ok := s.store.(diagnosisWorkflowStore)
	if !ok {
		writeError(w, 503, "UNAVAILABLE", "Asignación no disponible.")
		return
	}
	var v domain.DiagnosisAssignment
	if !decode(w, r, &v) {
		return
	}
	out, err := store.AssignDiagnosis(r.Context(), id, v, u)
	s.result(w, 200, out, err)
}
func (s *server) updateDiagnosis(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	store, ok := s.store.(diagnosisWorkflowStore)
	if !ok {
		writeError(w, 503, "UNAVAILABLE", "Diagnóstico no disponible.")
		return
	}
	var v domain.DiagnosisInput
	if !decode(w, r, &v) {
		return
	}
	out, err := store.SaveDiagnosis(r.Context(), id, v, u)
	s.result(w, 200, out, err)
}
