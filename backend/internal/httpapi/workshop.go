package httpapi

import (
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"net/http"
	"strconv"
)

func (s *server) workshopEndpoints() []endpoint {
	var out []endpoint
	for _, spec := range []struct{ resource, permission string }{{"recepciones", "recepciones.ver"}, {"diagnosticos", "diagnosticos.ver"}, {"cotizaciones", "cotizaciones.ver"}, {"ordenes", "ordenes.ver"}, {"servicios", "cotizaciones.ver"}, {"repuestos", "cotizaciones.ver"}, {"mecanicos", "mecanicos.ver"}, {"asignaciones", "mecanicos.ver"}, {"tareas", "auxilio.ver"}} {
		path := "/api/v1/" + spec.resource
		if spec.resource == "tareas" {
			path = "/api/v1/me/tareas"
		}
		out = append(out, endpoint{"GET", path, spec.permission, s.workshopList(spec.resource)})
	}
	for _, resource := range []string{"recepciones", "diagnosticos", "cotizaciones", "ordenes"} {
		out = append(out, endpoint{"GET", "/api/v1/" + resource + "/{id}", resource + ".ver", s.workshopItem(resource)})
	}
	out = append(out,
		endpoint{"PATCH", "/api/v1/ordenes/{id}/pagos/{paymentId}", "pagos.rectificar", s.reversePayment},
		endpoint{"GET", "/api/v1/cotizaciones/atenciones", "cotizaciones.ver", s.workshopList("atenciones")},
		endpoint{"GET", "/api/v1/me/diagnosticos", "diagnosticos.ver", s.diagnosisQueue(false)},
		endpoint{"GET", "/api/v1/diagnosticos/pendientes-cotizacion", "diagnosticos.ver", s.diagnosisQueue(true)},
		endpoint{"PATCH", "/api/v1/recepciones/{id}/mecanico", "diagnosticos.asignar", s.assignDiagnosis},
		endpoint{"PUT", "/api/v1/diagnosticos/{id}", "diagnosticos.crear", s.updateDiagnosis},
		endpoint{"GET", "/api/v1/ordenes/{id}/control-tecnico", "ordenes.ver", s.inspection},
		endpoint{"PUT", "/api/v1/ordenes/{id}/control-tecnico", "ordenes.editar", s.inspection},
		endpoint{"GET", "/api/v1/ordenes/{id}/ficha-tecnica.pdf", "pagos.ver", s.sheet},
		endpoint{"POST", "/api/v1/recepciones", "recepciones.crear", s.createReception},
		endpoint{"POST", "/api/v1/diagnosticos", "diagnosticos.crear", s.createDiagnosis},
		endpoint{"POST", "/api/v1/cotizaciones", "cotizaciones.crear", s.saveQuote},
		endpoint{"PUT", "/api/v1/cotizaciones/{id}", "cotizaciones.editar", s.saveQuote},
		endpoint{"POST", "/api/v1/ordenes", "ordenes.crear", s.createOrder},
		endpoint{"PATCH", "/api/v1/ordenes/{id}", "ordenes.editar", s.updateOrder},
		endpoint{"DELETE", "/api/v1/ordenes/{id}", "ordenes.eliminar", s.deleteOrder},
		endpoint{"GET", "/api/v1/ordenes/{id}/pagos", "pagos.ver", s.workshopItem("pagos")},
		endpoint{"POST", "/api/v1/ordenes/{id}/pagos", "pagos.registrar", s.registerPayment},
		endpoint{"GET", "/api/v1/ordenes/{id}/entrega", "pagos.ver", s.workshopItem("entrega")},
		endpoint{"POST", "/api/v1/ordenes/{id}/entrega", "pagos.entregar", s.deliver},
	)
	return out
}
func workshopFilter(r *http.Request, resource string) (domain.WorkshopFilter, error) {
	p, err := page(r)
	if err != nil {
		return domain.WorkshopFilter{}, err
	}
	f := domain.WorkshopFilter{Page: p}
	q := r.URL.Query()
	switch resource {
	case "atenciones":
		f.ParentID = q.Get("recepcionId")
	case "diagnosticos":
		f.ParentID = q.Get("recepcionId")
	case "cotizaciones":
		f.ParentID = q.Get("diagnosticoId")
	case "asignaciones":
		f.ParentID = q.Get("ordenId")
		f.ClientID = q.Get("usuarioId")
	}
	for field, value := range map[string]string{"parentId": f.ParentID, "usuarioId": f.ClientID} {
		if value != "" {
			if _, err := domain.ID(value); err != nil {
				return f, domain.Invalid(field, "Filtro de ID inválido.")
			}
		}
	}
	for key, dst := range map[string]*bool{"pendienteDiagnostico": &f.Pending, "sinOrden": &f.WithoutOrder} {
		if q.Get(key) != "" {
			b, err := strconv.ParseBool(q.Get(key))
			if err != nil {
				return f, domain.Invalid(key, "Filtro booleano inválido.")
			}
			*dst = b
		}
	}
	if state := q.Get("estado"); state != "" {
		if resource == "cotizaciones" {
			f.State, err = domain.QuoteState(state)
		} else if resource == "ordenes" || resource == "tareas" {
			f.State, err = domain.OrderState(state)
		} else {
			return f, domain.Invalid("estado", "Este módulo no admite filtro de estado.")
		}
	}
	return f, err
}
func (s *server) workshopList(resource string) func(http.ResponseWriter, *http.Request, domain.User) {
	return func(w http.ResponseWriter, r *http.Request, u domain.User) {
		f, err := workshopFilter(r, resource)
		if err != nil {
			s.failure(w, err)
			return
		}
		out, err := s.store.WorkshopList(r.Context(), resource, f, u)
		if err != nil {
			s.failure(w, err)
			return
		}
		respond(w, 200, out)
	}
}
func (s *server) workshopItem(resource string) func(http.ResponseWriter, *http.Request, domain.User) {
	return func(w http.ResponseWriter, r *http.Request, u domain.User) {
		id, ok := pathID(w, r)
		if !ok {
			return
		}
		out, err := s.store.WorkshopItem(r.Context(), resource, id, u)
		s.result(w, 200, out, err)
	}
}
func (s *server) createReception(w http.ResponseWriter, r *http.Request, u domain.User) {
	var v domain.ReceptionInput
	if !decode(w, r, &v) {
		return
	}
	out, err := s.store.CreateReception(r.Context(), v, u)
	s.result(w, 201, out, err)
}
func (s *server) createDiagnosis(w http.ResponseWriter, r *http.Request, u domain.User) {
	var v domain.DiagnosisInput
	if !decode(w, r, &v) {
		return
	}
	out, err := s.store.CreateDiagnosis(r.Context(), v, u)
	s.result(w, 201, out, err)
}
func (s *server) saveQuote(w http.ResponseWriter, r *http.Request, u domain.User) {
	id := ""
	status := 201
	if r.Method == "PUT" {
		var ok bool
		id, ok = pathID(w, r)
		if !ok {
			return
		}
		status = 200
	}
	var v domain.QuoteInput
	if !decodeLimit(w, r, &v, 1<<20) {
		return
	}
	out, err := s.store.SaveQuote(r.Context(), id, v, u)
	s.result(w, status, out, err)
}
func (s *server) createOrder(w http.ResponseWriter, r *http.Request, u domain.User) {
	var v domain.CreateOrderInput
	if !decode(w, r, &v) {
		return
	}
	out, err := s.store.CreateOrder(r.Context(), v, u)
	s.result(w, 201, out, err)
}
func (s *server) updateOrder(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var v domain.UpdateOrderInput
	if !decode(w, r, &v) {
		return
	}
	out, err := s.store.UpdateOrder(r.Context(), id, v, u)
	s.result(w, 200, out, err)
}
func (s *server) deleteOrder(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := s.store.DeleteOrder(r.Context(), id, u); err != nil {
		s.failure(w, err)
		return
	}
	w.WriteHeader(204)
}
func (s *server) registerPayment(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var v domain.PaymentInput
	if !decode(w, r, &v) {
		return
	}
	out, err := s.store.RegisterPayment(r.Context(), id, v, u)
	s.result(w, 201, out, err)
}
func (s *server) deliver(w http.ResponseWriter, r *http.Request, u domain.User) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var v domain.DeliveryInput
	if !decode(w, r, &v) {
		return
	}
	out, err := s.store.Deliver(r.Context(), id, v, u)
	s.result(w, 201, out, err)
}
