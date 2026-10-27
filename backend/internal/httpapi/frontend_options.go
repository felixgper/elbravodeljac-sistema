package httpapi

import (
	"context"
	"encoding/json"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"net/http"
)

type frontendOptionsStore interface {
	FrontendOptions(context.Context, string, string, domain.Page, domain.User) (any, error)
}

// Narrow projections for related form fields, without granting access to CRUD
// routes of other modules. PostgreSQL applies the order assignment scope.
func (s *server) frontendOptions(w http.ResponseWriter, r *http.Request, u domain.User) {
	module, resource := r.URL.Query().Get("module"), r.URL.Query().Get("resource")
	allowed := map[string]string{
		"clientes": "vehiculos", "vehiculos": "clientes",
		"recepciones":  "clientes vehiculos diagnosticos",
		"diagnosticos": "clientes vehiculos recepciones mecanicos",
		"cotizaciones": "clientes vehiculos recepciones diagnosticos ordenes",
		"ordenes":      "clientes vehiculos recepciones diagnosticos cotizaciones mecanicos",
		"pagos":        "clientes vehiculos recepciones diagnosticos cotizaciones ordenes mecanicos contactos",
		"auxilio":      "clientes vehiculos recepciones diagnosticos cotizaciones ordenes mecanicos",
		"mecanicos":    "clientes vehiculos recepciones diagnosticos cotizaciones ordenes",
	}
	valid := false
	for _, v := range splitWords(allowed[module]) {
		if v == resource {
			valid = true
		}
	}
	if !valid || !u.Can(module+".ver") {
		writeError(w, 403, "FORBIDDEN", "Consulta auxiliar no autorizada.")
		return
	}
	p, err := page(r)
	if err != nil {
		s.failure(w, err)
		return
	}
	if module == "pagos" && resource == "contactos" {
		id := r.URL.Query().Get("orderId")
		if _, err = domain.ID(id); err != nil {
			s.failure(w, err)
			return
		}
		o, err := s.store.WorkshopItem(r.Context(), "ordenes", id, u)
		if err != nil {
			s.failure(w, err)
			return
		}
		raw, err := json.Marshal(o)
		if err != nil {
			s.failure(w, err)
			return
		}
		var v struct {
			ClientID string `json:"clientId"`
		}
		if err = json.Unmarshal(raw, &v); err != nil {
			s.failure(w, err)
			return
		}
		contacts, err := s.store.Contacts(r.Context(), v.ClientID)
		if err != nil {
			s.failure(w, err)
			return
		}
		// Contacts are bounded by their client. Match the standard list envelope.
		start := (p.Number - 1) * p.Size
		end := start + p.Size
		total := len(contacts)
		if start > total {
			start = total
		}
		if end > total {
			end = total
		}
		respond(w, 200, domain.List[domain.Contact]{Data: contacts[start:end], Meta: domain.Meta{Page: p.Number, PageSize: p.Size, Total: total}})
		return
	}
	store, ok := s.store.(frontendOptionsStore)
	if !ok {
		writeError(w, 503, "UNAVAILABLE", "Consulta auxiliar no disponible.")
		return
	}
	out, err := store.FrontendOptions(r.Context(), module, resource, p, u)
	if err != nil {
		s.failure(w, err)
		return
	}
	respond(w, 200, out)
}

func splitWords(v string) []string {
	var out []string
	start := 0
	for i := 0; i <= len(v); i++ {
		if i == len(v) || v[i] == ' ' {
			if i > start {
				out = append(out, v[start:i])
			}
			start = i + 1
		}
	}
	return out
}
