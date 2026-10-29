package httpapi

import (
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"net/http"
	"time"
	_ "time/tzdata"
)

func (s *server) relationalEndpoints() []endpoint {
	out := []endpoint{}
	for _, spec := range []struct{ resource, perm string }{{"empleados", "usuarios.ver"}, {"usuarios", "usuarios.ver"}, {"roles", "usuarios.ver"}, {"permisos", "usuarios.ver"}, {"auxilios", "auxilio.ver"}, {"proveedores", "proveedores.ver"}, {"unidades-auxilio", "auxilio.ver"}} {
		out = append(out, endpoint{"GET", "/api/v1/" + spec.resource, spec.perm, s.relationalList(spec.resource)})
	}
	for _, spec := range []struct{ resource, perm string }{{"empleados", "usuarios.ver"}, {"usuarios", "usuarios.ver"}, {"auxilios", "auxilio.ver"}, {"proveedores", "proveedores.ver"}} {
		out = append(out, endpoint{"GET", "/api/v1/" + spec.resource + "/{id}", spec.perm, s.relationalItem(spec.resource)})
	}
	out = append(out,
		endpoint{"POST", "/api/v1/empleados", "usuarios.crear", s.saveEmployee}, endpoint{"PATCH", "/api/v1/empleados/{id}", "usuarios.editar", s.saveEmployee},
		endpoint{"POST", "/api/v1/usuarios", "usuarios.crear", s.saveUser}, endpoint{"PATCH", "/api/v1/usuarios/{id}", "usuarios.editar", s.saveUser}, endpoint{"DELETE", "/api/v1/usuarios/{id}", "usuarios.eliminar", s.deleteUser},
		endpoint{"GET", "/api/v1/roles/{id}/permisos", "usuarios.ver", s.relationalItem("rolePermissions")}, endpoint{"PUT", "/api/v1/roles/{id}/permisos", "usuarios.roles", s.rolePermissions}, endpoint{"POST", "/api/v1/roles/{id}/permisos/restablecer", "usuarios.roles", s.rolePermissions},
		endpoint{"POST", "/api/v1/auxilios", "auxilio.crear", s.saveAuxilio}, endpoint{"PATCH", "/api/v1/auxilios/{id}", "auxilio.crear", s.saveAuxilio},
		endpoint{"POST", "/api/v1/proveedores", "proveedores.crear", s.saveProvider}, endpoint{"PATCH", "/api/v1/proveedores/{id}", "proveedores.editar", s.saveProvider}, endpoint{"DELETE", "/api/v1/proveedores/{id}", "proveedores.eliminar", s.deleteProvider},
		endpoint{"GET", "/api/v1/proveedores/{id}/repuestos", "proveedores.ver", s.relationalList("providerParts")}, endpoint{"PUT", "/api/v1/proveedores/{id}/repuestos/{repuestoId}", "proveedores.editar", s.linkProviderPart}, endpoint{"DELETE", "/api/v1/proveedores/{id}/repuestos/{repuestoId}", "proveedores.editar", s.linkProviderPart},
		endpoint{"POST", "/api/v1/servicios", "cotizaciones.editar", s.saveCatalog("servicios")}, endpoint{"POST", "/api/v1/repuestos", "cotizaciones.editar", s.saveCatalog("repuestos")}, endpoint{"POST", "/api/v1/unidades-auxilio", "auxilio.crear", s.saveCatalog("unidades-auxilio")},
		endpoint{"GET", "/api/v1/dashboard", "inicio.ver", s.statistics(true)}, endpoint{"GET", "/api/v1/reportes", "reportes.ver", s.statistics(false)},
	)
	return out
}
func (s *server) relationalList(resource string) func(http.ResponseWriter, *http.Request, domain.User) {
	return func(w http.ResponseWriter, r *http.Request, u domain.User) {
		p, err := page(r)
		if err != nil {
			s.failure(w, err)
			return
		}
		parent := ""
		if resource == "providerParts" {
			var ok bool
			parent, ok = pathID(w, r)
			if !ok {
				return
			}
		}
		out, err := s.store.RelationalList(r.Context(), resource, parent, p)
		if err != nil {
			s.failure(w, err)
			return
		}
		respond(w, 200, out)
	}
}
func (s *server) relationalItem(resource string) func(http.ResponseWriter, *http.Request, domain.User) {
	return func(w http.ResponseWriter, r *http.Request, u domain.User) {
		id, ok := pathID(w, r)
		if !ok {
			return
		}
		out, err := s.store.RelationalItem(r.Context(), resource, id)
		s.result(w, 200, out, err)
	}
}
func (s *server) saveUser(w http.ResponseWriter, r *http.Request, u domain.User) {
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
	var v domain.UserInput
	if !decode(w, r, &v) {
		return
	}
	out, err := s.store.SaveUser(r.Context(), id, v, u)
	if err == nil && id != "" && (v.Password != nil || v.Active != nil || v.RoleID != nil || v.RemoveRoleID != nil) {
		s.tokens.RevokeSubject(id)
	}
	s.result(w, status, out, err)
}
func (s *server) deleteUser(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := s.store.DeleteUser(r.Context(), id, u); err != nil {
		s.failure(w, err)
		return
	}
	s.tokens.RevokeSubject(id)
	w.WriteHeader(204)
}
func (s *server) rolePermissions(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var v domain.MatrixInput
	reset := r.Method == "POST"
	if !reset && !decode(w, r, &v) {
		return
	}
	out, err := s.store.SetRolePermissions(r.Context(), id, v, reset, u)
	s.result(w, 200, out, err)
}
func (s *server) saveProvider(w http.ResponseWriter, r *http.Request, u domain.User) {
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
	var v domain.ProviderInput
	if !decode(w, r, &v) {
		return
	}
	out, err := s.store.SaveProvider(r.Context(), id, v)
	s.result(w, status, out, err)
}
func (s *server) deleteProvider(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := s.store.DeleteProvider(r.Context(), id); err != nil {
		s.failure(w, err)
		return
	}
	w.WriteHeader(204)
}
func (s *server) saveCatalog(resource string) func(http.ResponseWriter, *http.Request, domain.User) {
	return func(w http.ResponseWriter, r *http.Request, u domain.User) {
		var v domain.CatalogInput
		if !decode(w, r, &v) {
			return
		}
		out, err := s.store.SaveCatalog(r.Context(), resource, v)
		s.result(w, 201, out, err)
	}
}
func (s *server) saveAuxilio(w http.ResponseWriter, r *http.Request, u domain.User) {
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
	var v domain.AuxilioInput
	if !decode(w, r, &v) {
		return
	}
	out, err := s.store.SaveAuxilio(r.Context(), id, v, u)
	s.result(w, status, out, err)
}
func (s *server) linkProviderPart(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	part := r.PathValue("repuestoId")
	if _, err := domain.ID(part); err != nil {
		s.failure(w, err)
		return
	}
	remove := r.Method == "DELETE"
	var v domain.PartLinkInput
	if !remove && !decode(w, r, &v) {
		return
	}
	out, err := s.store.LinkProviderPart(r.Context(), id, part, v, remove)
	if err == nil && remove {
		w.WriteHeader(204)
		return
	}
	s.result(w, 200, out, err)
}
func statsRange(r *http.Request, dashboard bool) (time.Time, time.Time, error) {
	loc, err := time.LoadLocation("America/Lima")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	now := time.Now().In(loc)
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	end := start.AddDate(0, 1, 0)
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "Mes"
	}
	if !dashboard && period == "Este mes" {
		period = "Mes"
	}
	if !dashboard && period == "Últimos 6 meses" {
		start = start.AddDate(0, -5, 0)
	} else {
		start, end, err = businessPeriod(now, period)
		if err != nil {
			return start, end, err
		}
	}
	if dashboard {
		return start, end, nil
	}
	from, until := r.URL.Query().Get("desde"), r.URL.Query().Get("hasta")
	if from != "" || until != "" {
		if from == "" || until == "" {
			return start, end, domain.Invalid("period", "Envía desde y hasta juntos.")
		}
		start, err = time.Parse(time.RFC3339, from)
		if err != nil {
			return start, end, domain.Invalid("desde", "Usa RFC3339 con zona horaria.")
		}
		end, err = time.Parse(time.RFC3339, until)
		if err != nil {
			return start, end, domain.Invalid("hasta", "Usa RFC3339 con zona horaria.")
		}
	}
	if !end.After(start) || end.Sub(start) > 732*24*time.Hour {
		return start, end, domain.Invalid("period", "Intervalo positivo de máximo dos años; hasta es exclusivo.")
	}
	return start, end, nil
}
func (s *server) statistics(dashboard bool) func(http.ResponseWriter, *http.Request, domain.User) {
	return func(w http.ResponseWriter, r *http.Request, u domain.User) {
		start, end, err := statsRange(r, dashboard)
		if err != nil {
			s.failure(w, err)
			return
		}
		out, err := s.store.Statistics(r.Context(), dashboard, start, end)
		s.result(w, 200, out, err)
	}
}

// Half-open business periods in Lima; weeks start Monday, independent of host TZ.
func businessPeriod(now time.Time, period string) (time.Time, time.Time, error) {
	loc, err := time.LoadLocation("America/Lima")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	now = now.In(loc)
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	switch period {
	case "Semana":
		start = start.AddDate(0, 0, -(int(start.Weekday())+6)%7)
		return start, start.AddDate(0, 0, 7), nil
	case "Mes":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		return start, start.AddDate(0, 1, 0), nil
	case "Año":
		start = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, loc)
		return start, start.AddDate(1, 0, 0), nil
	}
	return start, start, domain.Invalid("period", "Selecciona Semana, Mes o Año.")
}
