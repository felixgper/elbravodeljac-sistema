package httpapi

import (
	"context"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"net/http"
)

type employeeStore interface {
	SaveEmployee(context.Context, string, domain.EmployeeInput, domain.User) (any, error)
}

func (s *server) saveEmployee(w http.ResponseWriter, r *http.Request, u domain.User) {
	id := ""
	status := 201
	if r.Method == "PATCH" {
		var ok bool
		id, ok = pathID(w, r)
		if !ok {
			return
		}
		status = 200
	}
	var v domain.EmployeeInput
	if !decode(w, r, &v) {
		return
	}
	store, ok := s.store.(employeeStore)
	if !ok {
		writeError(w, 503, "UNAVAILABLE", "Empleados no disponible.")
		return
	}
	out, err := store.SaveEmployee(r.Context(), id, v, u)
	s.result(w, status, out, err)
}
