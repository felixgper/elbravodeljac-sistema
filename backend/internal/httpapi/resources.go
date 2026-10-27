package httpapi

import (
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"net/http"
)

func (s *server) endpoints() []endpoint {
	return []endpoint{
		{"GET", "/api/v1/auth/me", "", func(http.ResponseWriter, *http.Request, domain.User) {}},
		{"POST", "/api/v1/auth/logout", "", func(http.ResponseWriter, *http.Request, domain.User) {}},
		{"GET", "/api/v1/clientes", "clientes.ver", s.clients},
		{"POST", "/api/v1/clientes", "clientes.crear", s.createClient},
		{"GET", "/api/v1/clientes/{id}", "clientes.ver", s.client},
		{"PATCH", "/api/v1/clientes/{id}", "clientes.editar", s.patchClient},
		{"DELETE", "/api/v1/clientes/{id}", "clientes.eliminar", s.deleteClient},
		{"GET", "/api/v1/clientes/{id}/contactos", "clientes.ver", s.contacts},
		{"POST", "/api/v1/clientes/{id}/contactos", "pagos.entregar", s.createContact},
		{"GET", "/api/v1/vehiculos", "vehiculos.ver", s.vehicles},
		{"POST", "/api/v1/vehiculos", "vehiculos.crear", s.createVehicle},
		{"GET", "/api/v1/vehiculos/{id}", "vehiculos.ver", s.vehicle},
		{"PATCH", "/api/v1/vehiculos/{id}", "vehiculos.editar", s.patchVehicle},
		{"DELETE", "/api/v1/vehiculos/{id}", "vehiculos.eliminar", s.deleteVehicle},
		{"POST", "/api/v1/vehiculos/{id}/transferencias", "vehiculos.editar", s.transferVehicle},
		{"GET", "/api/v1/vehiculos/{id}/historial", "vehiculos.ver", s.history},
	}
}
func (s *server) result(w http.ResponseWriter, status int, v any, err error) {
	if err != nil {
		s.failure(w, err)
		return
	}
	respond(w, status, map[string]any{"data": v})
}
func (s *server) clients(w http.ResponseWriter, r *http.Request, u domain.User) {
	p, err := page(r)
	if err != nil {
		s.failure(w, err)
		return
	}
	v, err := s.store.Clients(r.Context(), p)
	if err != nil {
		s.failure(w, err)
		return
	}
	respond(w, 200, v)
}
func (s *server) client(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	v, err := s.store.Client(r.Context(), id)
	s.result(w, 200, v, err)
}
func (s *server) createClient(w http.ResponseWriter, r *http.Request, u domain.User) {
	var v domain.ClientInput
	if !decode(w, r, &v) {
		return
	}
	if err := v.Validate(); err != nil {
		s.failure(w, err)
		return
	}
	out, err := s.store.CreateClient(r.Context(), v)
	if err == nil {
		w.Header().Set("Location", "/api/v1/clientes/"+out.ID)
	}
	s.result(w, 201, out, err)
}
func (s *server) patchClient(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var v map[string]string
	if !decode(w, r, &v) {
		return
	}
	if len(v) == 0 {
		s.failure(w, domain.Invalid("request", "Envía al menos un campo editable."))
		return
	}
	out, err := s.store.PatchClient(r.Context(), id, v)
	s.result(w, 200, out, err)
}
func (s *server) deleteClient(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := s.store.DeleteClient(r.Context(), id); err != nil {
		s.failure(w, err)
		return
	}
	w.WriteHeader(204)
}
func (s *server) contacts(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	v, err := s.store.Contacts(r.Context(), id)
	s.result(w, 200, v, err)
}
func (s *server) createContact(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var v domain.ContactInput
	if !decode(w, r, &v) {
		return
	}
	if err := v.Validate(); err != nil {
		s.failure(w, err)
		return
	}
	out, err := s.store.CreateContact(r.Context(), id, v)
	s.result(w, 201, out, err)
}
func (s *server) vehicles(w http.ResponseWriter, r *http.Request, u domain.User) {
	p, err := page(r)
	if err != nil {
		s.failure(w, err)
		return
	}
	v, err := s.store.Vehicles(r.Context(), p)
	if err != nil {
		s.failure(w, err)
		return
	}
	respond(w, 200, v)
}
func (s *server) vehicle(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	v, err := s.store.Vehicle(r.Context(), id)
	s.result(w, 200, v, err)
}
func (s *server) createVehicle(w http.ResponseWriter, r *http.Request, u domain.User) {
	var v domain.VehicleInput
	if !decode(w, r, &v) {
		return
	}
	if err := v.NormalizeValidate(); err != nil {
		s.failure(w, err)
		return
	}
	out, err := s.store.CreateVehicle(r.Context(), v, u.ID)
	if err == nil {
		w.Header().Set("Location", "/api/v1/vehiculos/"+out.ID)
	}
	s.result(w, 201, out, err)
}
func (s *server) patchVehicle(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var v domain.VehiclePatch
	if !decode(w, r, &v) {
		return
	}
	if v.Brand == nil && v.Model == nil && v.Year == nil && v.Mileage == nil && v.Notes == nil {
		s.failure(w, domain.Invalid("request", "Envía al menos un campo editable."))
		return
	}
	out, err := s.store.PatchVehicle(r.Context(), id, v)
	s.result(w, 200, out, err)
}
func (s *server) deleteVehicle(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := s.store.DeleteVehicle(r.Context(), id); err != nil {
		s.failure(w, err)
		return
	}
	w.WriteHeader(204)
}
func (s *server) transferVehicle(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var v domain.Transfer
	if !decode(w, r, &v) {
		return
	}
	if err := v.Validate(); err != nil {
		s.failure(w, err)
		return
	}
	out, err := s.store.TransferVehicle(r.Context(), id, v, u.ID)
	s.result(w, 200, out, err)
}
func (s *server) history(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	out, err := s.store.VehicleHistory(r.Context(), id)
	s.result(w, 200, out, err)
}
